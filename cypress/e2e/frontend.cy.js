const categories = ["environment", "politics", "economy", "technology", "crypto"]
const appOrigin = "http://localhost:8085"
const authOrigin = "http://localhost:8086"
const account = {
  email: `cypress-${Date.now()}@example.com`,
  password: "cypress-pass-123",
}

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

const waitForNewsFeed = () => {
  categories.forEach((category) => cy.wait(`@news-${category}`))
}

const registerAccount = () => {
  mockNewsFeed()
  cy.visit("/")
  cy.url().should("include", "localhost:8086")
  cy.contains("button", "Create account")
    .closest("form")
    .within(() => {
      cy.get('input[name="email"]').type(account.email)
      cy.get('input[name="password"]').type(account.password)
      cy.contains("button", "Create account").click()
    })

  cy.origin(appOrigin, { args: { appOrigin } }, ({ appOrigin }) => {
    cy.location("origin").should("eq", appOrigin)
  })
  waitForNewsFeed()
}

const loginToAccount = () => {
  mockNewsFeed()
  cy.visit("/")
  cy.url().should("include", "localhost:8086")
  cy.contains("button", "Sign in")
    .closest("form")
    .within(() => {
      cy.get('input[name="email"]').type(account.email)
      cy.get('input[name="password"]').type(account.password)
      cy.contains("button", "Sign in").click()
    })

  cy.origin(appOrigin, { args: { appOrigin } }, ({ appOrigin }) => {
    cy.location("origin").should("eq", appOrigin)
  })
  waitForNewsFeed()
}

describe("Bitcoin Trend Recon frontend", () => {
  it("creates a new account and redirects back to the app", () => {
    registerAccount()

    cy.origin(appOrigin, () => {
      cy.get('[data-cy="session-summary"]').should("contain.text", "Signed in as")
      cy.get('[data-cy="logout-link"]').should("have.attr", "href").and("contain", "/oauth2/sign_out?rd=")
    })
  })

  it("logs out through oauth2-proxy and returns to the auth page", () => {
    loginToAccount()

    cy.origin(appOrigin, () => {
      cy.get('[data-cy="session-summary"]').should("contain.text", "Signed in as")
      cy.get('[data-cy="logout-link"]').click()
    })

    cy.location("origin").should("eq", authOrigin)
    cy.contains("button", "Sign in").should("be.visible")
    cy.contains("button", "Create account").should("be.visible")
  })

  it("renders the aggregated news feed after login", () => {
    loginToAccount()

    cy.origin(appOrigin, () => {
      cy.get('[data-cy="news-status"]').should("have.text", "Loaded 5 items")
      cy.get('[data-cy="news-card"]').should("have.length", 5)
      cy.get('[data-cy="news-card"][data-category="crypto"]').within(() => {
        cy.get('[data-cy="news-title"]')
          .should("have.attr", "href", "https://example.com/crypto/5")
          .and("contain.text", "crypto headline 5")
      })
    })
  })

  it("persists the selected theme after reload for an authenticated user", () => {
    loginToAccount()

    cy.origin(appOrigin, () => {
      cy.get("body").should("have.attr", "data-theme", "dark")
      cy.get('[data-cy="theme-toggle"]').should("have.text", "Light").click()
      cy.get("body").should("have.attr", "data-theme", "light")
      cy.window().its("localStorage.theme").should("eq", "light")
    })

    mockNewsFeed()
    cy.visit("/")
    waitForNewsFeed()

    cy.origin(appOrigin, () => {
      cy.get("body").should("have.attr", "data-theme", "light")
      cy.get('[data-cy="theme-toggle"]').should("have.text", "Dark")
      cy.get('[data-cy="session-summary"]').should("contain.text", "Signed in as")
    })
  })

  it("sends selected categories and renders a successful prediction", () => {
    loginToAccount()

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

    cy.origin(appOrigin, () => {
      cy.get('[data-cy="category-button"][data-category="economy"]').click()
      cy.get('[data-cy="category-button"][data-category="crypto"]').click()
      cy.get('[data-cy="predict-button"]').click()
    })

    cy.wait("@predict").its("response.statusCode").should("eq", 200)
    cy.origin(appOrigin, () => {
      cy.get('[data-cy="prediction-status"]').should("have.text", "Prediction updated.")
      cy.get('[data-cy="prediction-result"]').within(() => {
        cy.contains("Target: 71234.5678")
        cy.contains("↗")
        cy.contains("⚡︎")
      })
    })
  })

  it("clicks predict and receives a non-empty 200 response", () => {
    loginToAccount()

    cy.intercept("POST", "/predict", {
      statusCode: 200,
      body: {
        target: 70123.4567,
        current: 70000,
        pred_horizon: 1,
      },
    }).as("predict")

    cy.origin(appOrigin, () => {
      cy.get('[data-cy="category-button"][data-category="technology"]').click()
      cy.get('[data-cy="predict-button"]').click()
    })

    cy.wait("@predict").then(({ response }) => {
      expect(response?.statusCode).to.equal(200)
      expect(response?.body).to.not.be.empty
    })

    cy.origin(appOrigin, () => {
      cy.get('[data-cy="prediction-result"]').invoke("text").should("not.be.empty")
    })
  })

  it("shows the backend error message when prediction fails", () => {
    loginToAccount()

    cy.intercept("POST", "/predict", {
      statusCode: 500,
      body: "LLM service unavailable",
    }).as("predict")

    cy.origin(appOrigin, () => {
      cy.get('[data-cy="category-button"][data-category="politics"]').click()
      cy.get('[data-cy="predict-button"]').click()
    })

    cy.wait("@predict").its("response.statusCode").should("eq", 500)
    cy.origin(appOrigin, () => {
      cy.get('[data-cy="prediction-result"]').should("have.text", "Prediction failed.")
      cy.get('[data-cy="prediction-status"]').should("have.text", "LLM service unavailable")
      cy.get('[data-cy="predict-button"]').should("not.be.disabled")
    })
  })
})
