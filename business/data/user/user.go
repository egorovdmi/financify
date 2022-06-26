package user

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/egorovdmi/financify/business/auth"
	"github.com/egorovdmi/financify/foundation/database"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("user not found")
	ErrInvalidID             = errors.New("ID is not in its proper form")
	ErrForbidden             = errors.New("authorization failed")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

type UserRepository struct {
	log *log.Logger
	db  *sqlx.DB
}

func NewUserRepository(log *log.Logger, db *sqlx.DB) UserRepository {
	return UserRepository{
		log: log,
		db:  db,
	}
}

func (r UserRepository) Create(ctx context.Context, traceID string, nu NewUser, now time.Time) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errors.Wrap(err, "generation password hash")
	}

	u := User{
		ID:           uuid.New().String(),
		Name:         nu.Name,
		Email:        nu.Email,
		PasswordHash: hash,
		Roles:        nu.Roles,
		DateCreated:  now.UTC(),
		DateUpdated:  now.UTC(),
	}

	const q = `INSERT INTO users
		(user_id, name, email, roles, password_hash, date_created, date_updated)
		VALUES($1, $2, $3, $4, $5, $6, $7)`

	r.log.Printf("%s : %s : query : %s", traceID, "UserRepository.Create",
		database.Log(q, u.ID, u.Name, u.Email, u.Roles, u.PasswordHash, u.DateCreated, u.DateUpdated))

	if _, err = r.db.ExecContext(ctx, q, u.ID, u.Name, u.Email, u.Roles, u.PasswordHash, u.DateCreated, u.DateUpdated); err != nil {
		return User{}, errors.Wrap(err, "inserting user")
	}

	return u, nil
}

func (r UserRepository) Update(ctx context.Context, traceID string, claims auth.Claims, userID string, uu UpdateUser, now time.Time) error {
	u, err := r.QueryByID(ctx, traceID, claims, userID)
	if err != nil {
		return err
	}

	if uu.Name != nil {
		u.Name = *uu.Name
	}
	if uu.Email != nil {
		u.Email = *uu.Email
	}
	if uu.Roles != nil {
		u.Roles = uu.Roles
	}
	if uu.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*uu.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.Wrap(err, "generation password hash")
		}
		u.PasswordHash = hash
	}
	u.DateUpdated = now.UTC()

	const q = `UPDATE users	SET 
		"name"=$2,
		"email"=$3,
		"roles"=$4,
		"password_hash"=$5,
		"date_updated"=$6
		WHERE user_id=$1`

	r.log.Printf("%s : %s : query : %s", traceID, "UserRepository.Update",
		database.Log(q, u.ID, u.Name, u.Email, u.Roles, u.PasswordHash, u.DateUpdated))

	if _, err = r.db.ExecContext(ctx, q, u.ID, u.Name, u.Email, u.Roles, u.PasswordHash, u.DateUpdated); err != nil {
		return errors.Wrap(err, "updating user")
	}

	return nil
}

func (r UserRepository) Delete(ctx context.Context, traceID string, userID string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return ErrInvalidID
	}

	const q = `DELETE FROM users WHERE user_id=$1`

	r.log.Printf("%s : %s : query : %s", traceID, "UserRepository.Delete",
		database.Log(q, userID))

	if _, err := r.db.ExecContext(ctx, q, userID); err != nil {
		return errors.Wrap(err, "deleting user")
	}

	return nil
}

func (r UserRepository) Query(ctx context.Context, traceID string) ([]User, error) {
	const q = `SELECT * FROM users`

	r.log.Printf("%s : %s : query : %s", traceID, "UserRepository.Query",
		database.Log(q))

	users := []User{}
	if err := r.db.SelectContext(ctx, &users, q); err != nil {
		return nil, errors.Wrap(err, "selecting users")
	}

	return users, nil
}

func (r UserRepository) QueryByID(ctx context.Context, traceID string, claims auth.Claims, userID string) (User, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return User{}, ErrInvalidID
	}

	if !claims.Authorize(auth.RoleAdmin) && claims.Subject != userID {
		return User{}, ErrForbidden
	}

	const q = `SELECT * FROM users WHERE user_id=$1`

	r.log.Printf("%s : %s : query : %s", traceID, "UserRepository.QueryByID",
		database.Log(q, userID))

	var u User
	if err := r.db.GetContext(ctx, &u, q, userID); err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrNotFound
		}
		return User{}, errors.Wrapf(err, "selecting user %q", userID)
	}

	return u, nil
}

func (r UserRepository) QueryByEmail(ctx context.Context, traceID string, claims auth.Claims, email string) (User, error) {
	const q = `SELECT * FROM users WHERE email=$1`

	r.log.Printf("%s : %s : query : %s", traceID, "UserRepository.QueryByEmail",
		database.Log(q, email))

	var u User
	if err := r.db.GetContext(ctx, &u, q, email); err != nil {
		if err == sql.ErrNoRows {
			return User{}, ErrNotFound
		}
		return User{}, errors.Wrapf(err, "selecting user %q", email)
	}

	if !claims.Authorize(auth.RoleAdmin) && claims.Subject != u.ID {
		return User{}, ErrForbidden
	}

	return u, nil
}

func (r UserRepository) Authenticate(ctx context.Context, traceID string, email string, password string, now time.Time) (auth.Claims, error) {
	adminClaims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "service project",
			Audience:  jwt.ClaimStrings{"students"},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Roles: []string{auth.RoleAdmin},
	}

	usr, err := r.QueryByEmail(ctx, traceID, adminClaims, email)
	if err != nil {
		switch err {
		case ErrNotFound:
			return auth.Claims{}, ErrAuthenticationFailure
		case ErrForbidden:
			return auth.Claims{}, ErrAuthenticationFailure
		default:
			return auth.Claims{}, errors.Wrap(err, "unable to query user by email")
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return auth.Claims{}, errors.Wrap(err, "generation password hash")
	}

	if len(usr.PasswordHash) != len(hash) {
		return auth.Claims{}, ErrAuthenticationFailure
	}

	for i, b := range usr.PasswordHash {
		if b != hash[i] {
			return auth.Claims{}, ErrAuthenticationFailure
		}
	}

	claims := auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "service project",
			Audience:  jwt.ClaimStrings{"students"},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		Roles: []string{auth.RoleUser},
	}

	return claims, nil
}
