-- +goose Up
INSERT INTO operations (operation_type) VALUES 
('+'),
('-'),
('*'),
('/');

-- +goose Down
DELETE FROM operations WHERE operation_type = '+';
DELETE FROM operations WHERE operation_type = '-';
DELETE FROM operations WHERE operation_type = '/';
DELETE FROM operations WHERE operation_type = '*';