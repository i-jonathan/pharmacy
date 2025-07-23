package repository

const createUserQuery = `INSERT INTO users (username, password) VALUES ($1, $2)`
const usernameExistsQuery = `SELECT 1 FROM users WHERE username = $1 LIMIT 1`
const fetchUserByNameQuery = `SELECT id, username, password FROM users WHERE username = $1`