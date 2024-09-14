-- name: GetDeviceInfo :one
SELECT * FROM `users` WHERE username = ? LIMIT 1;



