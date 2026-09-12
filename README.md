# Go Media Server

## 1. Edit config.json

```json
{
  "username": "admin",
  "password": "password",
  "port": "8080",
  "libraries": [
    {
      "name": "My Media",
      "path": "C:/Users/you/Videos"
    }
  ]
}
```

- **username / password** — required to open the site at all (browser's
  built-in login popup, HTTP Basic Auth). Change these from the example
  values before you use this on anything but localhost. Leave both blank
  (`""`) to disable the login prompt entirely.
- **port** — just the number, e.g. `"8080"`. The server turns it into
  `:8080` automatically.
- **libraries** — one or more folders to browse. Each becomes a separate
  entry in the library dropdown at the top of the page.

**Important on Windows:** use forward slashes (`C:/Users/you/Videos`) or
doubled backslashes (`"C:\\Users\\you\\Videos"`) for paths. A single
backslash isn't valid inside a JSON string and the server will fail to
parse the file.

You can list more than one folder:

```json
{
  "username": "admin",
  "password": "password",
  "port": "8080",
  "libraries": [
    { "name": "Movies", "path": "D:/Media/Movies" },
    { "name": "TV Shows", "path": "D:/Media/TV" },
    { "name": "Photos", "path": "C:/Users/you/Pictures" }
  ]
}
```

Every subfolder inside a configured path is automatically browsable — you
don't need to list them individually.

## 2. Run it

From inside the `mediaserver` folder:

```powershell
go run .
```

## 3. Build a standalone .exe

```powershell
go build -o mediaserver.exe .
.\mediaserver.exe
```

Because the frontend is embedded with `go:embed`, `mediaserver.exe` is a
single file — copy it anywhere, keep `config.json` next to it (or set
`CONFIG_FILE` to point elsewhere), and run it.

## 4. Images, video, and audio together

`.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`, and `.bmp` are treated as media
files alongside video/audio, and show up in the same folder listing with
their own icon.

- **Desktop/laptop**: clicking an image displays it in the same player
  container the video uses — no separate popup, since there's already a
  large flexible area to show it in.
- **Mobile**: the video area is a fixed 16:9 box, which doesn't suit
  arbitrary photo aspect ratios, so images instead open in a full-screen
  lightbox (tap the ✕, tap outside the image, or press Esc to close it).

This split is decided by screen width at the moment you click (the same
800px breakpoint used for the sidebar/stacked layout switch), not by
device type, so resizing a desktop browser window narrow enough will also
switch images to the lightbox behavior.

## 5. How streaming/seeking works

`handleStreamMedia` opens the requested file and hands it to Go's
`http.ServeContent`, which:

- Sets `Content-Type` based on the file extension (registered explicitly in
  `init()` for common formats).
- Honors `Range` headers, so the browser can jump to any point in a video
  without downloading it from the start — this is what makes the `<video>`
  scrubber work, and lets images/audio load efficiently too.

## 6. Security notes

- Every request is resolved against the configured library root and
  checked so it can never escape that folder (no `..` traversal, symlinked
  folders aren't followed). Only the file types listed in `allowedExt` in
  `main.go` are ever listed or served.
- Authentication is plain HTTP Basic Auth: simple, and enough to keep
  casual access out on a home/trusted network. **Don't expose this server
  directly to the public internet**,
