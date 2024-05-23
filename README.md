# GoWebPractice

This is a learning log from the Let's Go! book by Alex Edwards.

### Directory Structure

- The `cmd` directory will contain the application-specific code for the executable applications in the project. For now we’ll have just one executable application — the web application — which will live under the `cmd/web` directory.

- The `internal` directory will contain the ancillary non-application-specific code used in the project. We’ll use it to hold potentially reusable code like `validation` `helpers` and the `SQL database models` for the project.

- The `ui` directory will contain the user-interface assets used by the web application. Specifically, the `ui/html` directory will contain `HTML templates`, and the `ui/static` directory will contain static files (like `CSS` and `images`).

### Why using this structure?

- It is very scalable if you want to add another executable application to your project.
  - For example, you may want to add a CLI (Command Line Interface) to automate some future administrative tasks. With this structure, you can create this CLI application under `cmd/cli`, and it will be able to import and reuse all the code you have written under the `internal` directory.
- It prevents other libraries from importing and relying on packages in our `internal` directory.
  - The directory name `Internal` has a special meaning and behaviour in Go: any packages located in that directory can only be imported through code in the parent directory of the `internal` directory. Any packages under `Internal` cannot be imported by code external to our project.

#### request processing flow

```shell=
middleware(1) → middleware(2) → servemux → application handler
```

## DB schema

```sql=

+---------+--------------+------+-----+---------+----------------+
| Field   | Type         | Null | Key | Default | Extra          |
+---------+--------------+------+-----+---------+----------------+
| id      | int          | NO   | PRI | NULL    | auto_increment |
| title   | varchar(100) | NO   |     | NULL    |                |
| content | text         | NO   |     | NULL    |                |
| created | datetime     | NO   | MUL | NULL    |                |
| expires | datetime     | NO   |     | NULL    |                |
+---------+--------------+------+-----+---------+----------------+


CREATE DATABASE snippetbox CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

CREATE TABLE snippets (
  id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
  title VARCHAR(100) NOT NULL,
  content TEXT NOT NULL,
  created DATETIME NOT NULL,
  expires DATETIME NOT NULL
);

CREATE INDEX idx_snippets_created ON snippets(created);


+--------+--------------+------+-----+---------+-------+
| Field  | Type         | Null | Key | Default | Extra |
+--------+--------------+------+-----+---------+-------+
| token  | char(43)     | NO   | PRI | NULL    |       |
| data   | blob         | NO   |     | NULL    |       |
| expiry | timestamp(6) | NO   | MUL | NULL    |       |
+--------+--------------+------+-----+---------+-------+


CREATE TABLE sessions (
  token CHAR(43) PRIMARY KEY,
  data BLOB NOT NULL,
  expiry TIMESTAMP(6) NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);


+-----------------+--------------+------+-----+---------+----------------+
| Field           | Type         | Null | Key | Default | Extra          |
+-----------------+--------------+------+-----+---------+----------------+
| id              | int          | NO   | PRI | NULL    | auto_increment |
| name            | varchar(255) | NO   |     | NULL    |                |
| email           | varchar(255) | NO   | UNI | NULL    |                |
| hashed_password | char(60)     | NO   |     | NULL    |                |
| created         | datetime     | NO   |     | NULL    |                |
+-----------------+--------------+------+-----+---------+----------------+

CREATE TABLE users (
  id INTEGER NOT NULL PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL,
  hashed_password CHAR(60) NOT NULL,
  created DATETIME NOT NULL
);

ALTER TABLE users ADD CONSTRAINT users_uc_email UNIQUE (email);

```

#### API Spec

| Method | Pattern            | Handler           | Action                                         |
| ------ | ------------------ | ----------------- | ---------------------------------------------- |
| GET    | /                  | home              | Display the home page                          |
| GET    | /snippet/view/:id  | snippetView       | Display a specific snippet                     |
| GET    | /snippet/create    | snippetCreate     | Display a HTML form for creating a new snippet |
| POST   | /snippet/create    | snippetCreatePost | Create a new snippet                           |
| GET    | /user/signup       | userSignup        | Display a HTML form for signing up a new user  |
| POST   | /user/signup       | userSignupPost    | Create a new user                              |
| GET    | /user/login        | userLogin         | Display a HTML form for logging in a user      |
| POST   | /user/login        | userLoginPost     | Authenticate and login the user                |
| POST   | /user/logout       | userLogoutPost    | Logout the user                                |
| GET    | /static/\*filepath | http.FileServer   | Serve a specific static file                   |

#### SSL

```shell=
go run /opt/homebrew/Cellar/go/1.22.3/libexec/src/crypto/tls/generate_cert.go --rsa-bits=2048 --host=localhost

curl -i -X POST http://localhost:4000/snippet/view/1

```

```shell=
// Installation
brew services start mysql
// Adding a user for the first time with root login settings
sudo mysql
// Log in afterwards
mysql -D snippetbox -u web -p
```

```sql=
CREATE USER 'web'@'localhost';

GRANT SELECT, INSERT, UPDATE, DELETE ON snippetbox.* TO 'web'@'localhost';

ALTER USER 'web'@'localhost' IDENTIFIED BY 'pass';

```
