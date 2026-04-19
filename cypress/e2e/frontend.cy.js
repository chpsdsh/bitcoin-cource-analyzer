const categories = ["environment", "politics", "economy", "technology", "crypto"]

const buildNewsItem = (category, index) => ({
  category,
  title: `${category} headline ${index}`,
  url: `https://example.com/${category}/${index}`,
  socialimage: `https://images.example.com/${category}/${index}.jpg`,
})

const mockNewsFeed = () => {
  categories.forEach((category, index) => {
    cy.intercept("GET", `/news/${category}`, {
      statusCode: 200,
      body: {
        news: [buildNewsItem(category, index + 1)],
      },
    }).as(`news-${category}`)
  })
}

const mockSession = () => {
  cy.intercept("GET", "/session", {
    statusCode: 200,
    body: {
      user: "btc-operator",
      email: "trader@example.com",
    },
  }).as("session")
}

const visitApp = () => {
  mockSession()
  mockNewsFeed()
  cy.visit("/")
  cy.wait("@session")
  categories.forEach((category) => cy.wait(`@news-${category}`))
}

describe("Bitcoin Trend Recon frontend", () => {
  it("renders the aggregated news feed on first load", () => {
    visitApp()

    cy.get('[data-cy="news-status"]').should("have.text", "Loaded 5 items")
    cy.get('[data-cy="news-card"]').should("have.length", 5)
    cy.get('[data-cy="news-card"][data-category="crypto"]').within(() => {
      cy.get('[data-cy="news-title"]')
        .should("have.attr", "href", "https://example.com/crypto/5")
        .and("contain.text", "crypto headline 5")
    })
  })

  it("persists the selected theme after reload", () => {
    visitApp()

    cy.get("body").should("have.attr", "data-theme", "dark")
    cy.get('[data-cy="theme-toggle"]').should("have.text", "Light").click()
    cy.get("body").should("have.attr", "data-theme", "light")
    cy.window().its("localStorage.theme").should("eq", "light")

    mockSession()
    mockNewsFeed()
    cy.reload()
    cy.wait("@session")
    categories.forEach((category) => cy.wait(`@news-${category}`))

    cy.get("body").should("have.attr", "data-theme", "light")
    cy.get('[data-cy="theme-toggle"]').should("have.text", "Dark")
  })

  it("shows the authenticated identity and logout entrypoint", () => {
    visitApp()

    cy.get('[data-cy="session-summary"]').should("contain.text", "Signed in as trader@example.com")
    cy.get('[data-cy="logout-link"]').should("have.attr", "href").and("contain", "/oauth2/sign_out?rd=")
  })

  it("sends selected categories and renders a successful prediction", () => {
    visitApp()

    cy.intercept("POST", "/predict", (request) => {
      expect(request.body).to.deep.equal({
        categoriesList: ["economy", "crypto"],
      })

      request.reply({
        statusCode: 200,
        body: {
          target: 71234.5678,
          current: 70000,
          pred_horizon: 2,
        },
      })
    }).as("predict")

    cy.get('[data-cy="category-button"][data-category="economy"]').click()
    cy.get('[data-cy="category-button"][data-category="crypto"]').click()
    cy.get('[data-cy="predict-button"]').click()

    cy.wait("@predict")
    cy.get('[data-cy="prediction-status"]').should("have.text", "Prediction updated.")
    cy.get('[data-cy="prediction-result"]').within(() => {
      cy.contains("Target: 71234.5678")
      cy.contains("↗")
      cy.contains("⚡︎")
    })
  })

  it("clicks predict and receives a non-empty 200 response", () => {
    visitApp()

    cy.intercept("POST", "/predict", {
      statusCode: 200,
      body: {
        target: 70123.4567,
        current: 70000,
        pred_horizon: 1,
      },
    }).as("predict")

    cy.get('[data-cy="category-button"][data-category="technology"]').click()
    cy.get('[data-cy="predict-button"]').click()

    cy.wait("@predict").then(({ response }) => {
      expect(response?.statusCode).to.equal(200)
      expect(response?.body).to.not.be.empty
    })

    cy.get('[data-cy="prediction-result"]')
      .invoke("text")
      .should("not.be.empty")
  })

  it("shows the backend error message when prediction fails", () => {
    visitApp()

    cy.intercept("POST", "/predict", {
      statusCode: 500,
      body: "LLM service unavailable",
    }).as("predict")

    cy.get('[data-cy="category-button"][data-category="politics"]').click()
    cy.get('[data-cy="predict-button"]').click()

    cy.wait("@predict")
    cy.get('[data-cy="prediction-result"]').should("have.text", "Prediction failed.")
    cy.get('[data-cy="prediction-status"]').should("have.text", "LLM service unavailable")
    cy.get('[data-cy="predict-button"]').should("not.be.disabled")
  })
})
