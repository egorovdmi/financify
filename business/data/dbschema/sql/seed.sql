INSERT INTO users (user_id, name, email, roles, password_hash, date_created, date_updated) VALUES
	('5cf37266-3473-4006-984f-9325122678b7', 'Admin Gopher', 'admin@example.com', '{ADMIN,USER}', '$2a$10$gmZAzA.49G9DYSLXqeSt8.swqCJaB2EXgfQFSZJGFeovKwFrIi9gi', '2019-03-24 00:00:00', '2019-03-24 00:00:00'),
	('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'User Gopher', 'user@example.com', '{USER}', '$2a$10$gmZAzA.49G9DYSLXqeSt8.swqCJaB2EXgfQFSZJGFeovKwFrIi9gi', '2019-03-24 00:00:00', '2019-03-24 00:00:00')
	ON CONFLICT DO NOTHING;

INSERT INTO scopes (scope_id, user_id, title, amount, date_created, date_updated) VALUES
	('79ee821f-0a5b-4416-a77c-176cbfa14e4d', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'Домашняя бухгалтерия', 0, '2022-06-17 00:00:00', '2022-06-17 00:00:00')
	ON CONFLICT DO NOTHING;

INSERT INTO wallets (wallet_id, scope_id, user_id, title, amount, date_created, date_updated) VALUES
	('a11af2a9-9b3c-4950-bf8e-bf0d3c6399f2', '79ee821f-0a5b-4416-a77c-176cbfa14e4d', '45b5fbd3-755f-4379-8f07-a58d4a30fa2f', 'Наличные', 0, '2022-06-17 00:00:00', '2022-06-17 00:00:00')
	ON CONFLICT DO NOTHING;
