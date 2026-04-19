const { defineConfig } = require("cypress")

const host = "127.0.0.1"
const port = 8085

module.exports = defineConfig({
  experimentalWebKitSupport: true,
  e2e: {
    baseUrl: `http://${host}:${port}`,
  },
  video: false,
})
