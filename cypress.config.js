const { createReadStream, existsSync, statSync } = require("node:fs")
const http = require("node:http")
const path = require("node:path")
const { defineConfig } = require("cypress")

const host = "127.0.0.1"
const port = 8085
const frontendRoot = path.join(__dirname, "web-frontend", "html")

let server

const mimeTypes = {
  ".css": "text/css; charset=utf-8",
  ".html": "text/html; charset=utf-8",
  ".js": "application/javascript; charset=utf-8",
  ".json": "application/json; charset=utf-8",
  ".png": "image/png",
  ".svg": "image/svg+xml",
}

const sendIndex = (response) => {
  const indexPath = path.join(frontendRoot, "index.html")
  response.writeHead(200, { "Content-Type": "text/html; charset=utf-8" })
  createReadStream(indexPath).pipe(response)
}

const createServer = () =>
  http.createServer((request, response) => {
    const requestPath = new URL(request.url, `http://${host}:${port}`).pathname
    const normalizedPath = requestPath === "/" ? "/index.html" : requestPath
    const filePath = path.normalize(path.join(frontendRoot, normalizedPath))

    if (!filePath.startsWith(frontendRoot)) {
      response.writeHead(403)
      response.end("Forbidden")
      return
    }

    if (!existsSync(filePath) || statSync(filePath).isDirectory()) {
      sendIndex(response)
      return
    }

    const extension = path.extname(filePath)
    response.writeHead(200, {
      "Content-Type": mimeTypes[extension] || "application/octet-stream",
    })
    createReadStream(filePath).pipe(response)
  })

const ensureServer = async () => {
  if (server) {
    return
  }

  server = createServer()
  await new Promise((resolve) => server.listen(port, host, resolve))
}

const closeServer = async () => {
  if (!server) {
    return
  }

  await new Promise((resolve) => server.close(resolve))
  server = undefined
}

module.exports = defineConfig({
  experimentalWebKitSupport: true,
  e2e: {
    baseUrl: `http://${host}:${port}`,
    setupNodeEvents(on) {
      on("before:run", ensureServer)
      on("before:spec", ensureServer)
      on("after:run", closeServer)
    },
  },
  video: false,
})
