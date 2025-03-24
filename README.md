# GAuss

GAuss is a Google OAuth2 authentication service written in Go. It provides a simple, secure way to authenticate users with Google, store their session data, and optionally use a custom login template.

---

## Features

- **OAuth2** with Google
- **Session Management** using [gorilla/sessions](https://github.com/gorilla/sessions)
- **Embeddable Templates** for the login page (default or custom)
- **Dashboard** showing user information after login

---

## Getting Started

### Prerequisites

1. **Go** (version 1.23.4 or later recommended).
2. A **Google Cloud** project with OAuth credentials:
   - Client ID
   - Client Secret
3. A **session secret** (any random string or generated key).

### Environment Variables

Set the following environment variables before running GAuss:

- `GOOGLE_CLIENT_ID` – Your Google OAuth2 client ID.
- `GOOGLE_CLIENT_SECRET` – Your Google OAuth2 client secret.
- `SESSION_SECRET` – The secret key for signing sessions.

For example, you might place them in an `.env` file (excluded from version control):

```bash
GOOGLE_CLIENT_ID="your-client-id.apps.googleusercontent.com"
GOOGLE_CLIENT_SECRET="your-google-client-secret"
SESSION_SECRET="random-secret"
```

### Build and Run

1. **Clone** the repository or place the files in your Go workspace.
2. **Install** dependencies:
   ```bash
   go mod tidy
   ```
3. **Build** the binary:
   ```bash
   go build -o gauss cmd/web/main.go
   ```
4. **Run** the binary:
   ```bash
   ./gauss
   ```

GAuss will start listening on port `:8080`.

---

## Custom Login Template

You can override the default embedded `login.html` by passing the `--template` flag at runtime:

```bash
./gauss --template="/path/to/your/custom_login.html"
```

- If the flag is **not** provided, GAuss uses its default embedded `login.html`.
- If the flag **is** provided, GAuss will parse your custom file and replace the embedded `login.html`.

### Example

```bash
./gauss --template="templates/custom_login.html"
```

Ensure that your custom file exists and is accessible. Otherwise, you’ll get an error like `template: pattern matches no files`.

---

## Usage

1. **Set Environment Variables**: `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `SESSION_SECRET`.
2. **Run GAuss** (default):
   ```bash
   go run cmd/web/main.go
   ```
   or
   ```bash
   go build -o gauss cmd/web/main.go
   ./gauss
   ```
3. **Optional**: Pass a custom template, e.g.:
   ```bash
   ./gauss --template="templates/custom_login.html"
   ```
4. **Navigate** to [http://localhost:8080/](http://localhost:8080/) in your browser.
    - You will be redirected to the `/login` route if not logged in.
    - Clicking **Sign In with Google** starts the OAuth2 flow.
    - On success, you’ll land on the **Dashboard** (`/dashboard`).

---

## Routes

- **`/login`** – Displays the login page (`login.html` or your custom file).
- **`/auth/google`** – Initiates Google OAuth2 flow.
- **`/auth/google/callback`** – Google redirects here with an authorization code.
- **`/logout`** – Logs out the user by clearing session data.
- **`/dashboard`** – Protected route showing user info.

---

## Troubleshooting

1. **No custom file found**:  
   If you see `template: pattern matches no files`, ensure your custom template path is correct and accessible.
2. **State mismatch**:  
   If you see `error=invalid_state` in the URL, your session might have expired or your request was tampered with.
3. **Token exchange failed**:  
   Double-check your client ID and client secret, and that Google OAuth credentials are set correctly.

---

## License

This project does not specify a license by default. Add a LICENSE file if you plan to distribute or use this in production.

---

## Contributing

Feel free to open issues or pull requests. All contributions are welcome.

---

**Enjoy using GAuss for your Google OAuth2 authentication!**
