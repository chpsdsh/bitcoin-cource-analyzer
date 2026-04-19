const { defineConfig } = require("cypress")

const host = "localhost"
const port = 8085

module.exports = defineConfig({
  experimentalWebKitSupport: true,
  e2e: {
    baseUrl: `http://${host}:${port}`,
  },
  video: false,
})
