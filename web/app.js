const singleForm = document.querySelector("#single-form");
const compareForm = document.querySelector("#compare-form");
const singleResult = document.querySelector("#single-result");
const compareResult = document.querySelector("#compare-result");
const trendingList = document.querySelector("#trending-list");
const trendingState = document.querySelector("#trending-state");
const lastUpdated = document.querySelector("#last-updated");
const apiStatus = document.querySelector("#api-status");
const apiIndicator = document.querySelector("#api-indicator");
const refreshTrendingButton = document.querySelector("#refresh-trending");

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "USD",
  maximumFractionDigits: 2,
});

function escapeHTML(value) {
  return String(value).replace(/[&<>"']/g, (character) => {
    switch (character) {
      case "&":
        return "&amp;";
      case "<":
        return "&lt;";
      case ">":
        return "&gt;";
      case '"':
        return "&quot;";
      default:
        return "&#39;";
    }
  });
}

function normalizeCoinId(value) {
  return value.trim().toLowerCase();
}

function formatPrice(value) {
  if (!Number.isFinite(value)) {
    return "Unavailable";
  }

  if (value >= 1) {
    return currencyFormatter.format(value);
  }

  if (value >= 0.01) {
    return `$${value.toFixed(4)}`;
  }

  return `$${value.toFixed(8)}`;
}

function renderSingleState(title, body, valueText) {
  singleResult.innerHTML = `
    <p class="result-label">${escapeHTML(title)}</p>
    <h3>${valueText ? `<span class="result-value">${escapeHTML(valueText)}</span>` : escapeHTML(body)}</h3>
    ${valueText ? `<p class="result-note">${escapeHTML(body)}</p>` : '<p class="result-note">Live quotes are fetched from the backend API.</p>'}
  `;
}

function renderCompareState(title, content) {
  compareResult.innerHTML = `
    <p class="result-label">${escapeHTML(title)}</p>
    ${content}
  `;
}

function setConnectionState(text, mode) {
  apiStatus.textContent = text;
  apiIndicator.classList.remove("state-success", "state-error");

  if (mode) {
    apiIndicator.classList.add(mode);
  }
}

async function requestJSON(url) {
  const response = await fetch(url);
  let payload = {};

  try {
    payload = await response.json();
  } catch (error) {
    payload = {};
  }

  if (!response.ok) {
    throw new Error(payload.error || "Request failed");
  }

  return payload;
}

async function handleSingleLookup(event) {
  event.preventDefault();

  const symbol = normalizeCoinId(document.querySelector("#single-symbol").value);
  if (!symbol) {
    renderSingleState("Missing input", "Enter a CoinGecko coin ID like bitcoin or ethereum.");
    return;
  }

  renderSingleState("Loading", "Fetching the latest quote...");

  try {
    const data = await requestJSON(`/api/price?symbol=${encodeURIComponent(symbol)}`);
    renderSingleState(
      data.symbol.toUpperCase(),
      `Current USD quote for ${data.symbol}.`,
      formatPrice(data.price),
    );
    setConnectionState("Live data connected", "state-success");
  } catch (error) {
    renderSingleState("Unavailable", error.message);
    setConnectionState("Could not reach live data", "state-error");
  }
}

async function handleComparison(event) {
  event.preventDefault();

  const first = normalizeCoinId(document.querySelector("#first-symbol").value);
  const second = normalizeCoinId(document.querySelector("#second-symbol").value);

  if (!first || !second) {
    renderCompareState(
      "Missing input",
      "<h3>Both coin IDs are required.</h3><p class=\"result-note\">Try bitcoin against ethereum or solana.</p>",
    );
    return;
  }

  renderCompareState(
    "Loading",
    "<h3>Fetching both prices...</h3><p class=\"result-note\">This compares two live USD quotes from the backend.</p>",
  );

  try {
    const data = await requestJSON(
      `/api/prices?symbol1=${encodeURIComponent(first)}&symbol2=${encodeURIComponent(second)}`,
    );

    const ratio = data.price2 !== 0 ? `${(data.price1 / data.price2).toFixed(4)}x` : "n/a";

    renderCompareState(
      "Live comparison",
      `
        <h3>${escapeHTML(data.symbol1)} vs ${escapeHTML(data.symbol2)}</h3>
        <div class="result-grid">
          <div class="metric">
            <strong>${escapeHTML(data.symbol1)}</strong>
            <span>${escapeHTML(formatPrice(data.price1))}</span>
          </div>
          <div class="metric">
            <strong>${escapeHTML(data.symbol2)}</strong>
            <span>${escapeHTML(formatPrice(data.price2))}</span>
          </div>
          <div class="metric">
            <strong>Price ratio</strong>
            <span>${escapeHTML(ratio)}</span>
          </div>
          <div class="metric">
            <strong>Higher asset</strong>
            <span>${escapeHTML(data.price1 >= data.price2 ? data.symbol1 : data.symbol2)}</span>
          </div>
        </div>
      `,
    );
    setConnectionState("Live data connected", "state-success");
  } catch (error) {
    renderCompareState(
      "Unavailable",
      `<h3>Comparison failed.</h3><p class="result-note">${escapeHTML(error.message)}</p>`,
    );
    setConnectionState("Could not reach live data", "state-error");
  }
}

function renderTrending(coins) {
  if (!Array.isArray(coins) || coins.length === 0) {
    trendingList.innerHTML = `
      <div class="empty-state">
        <p>No trending coins are available right now.</p>
      </div>
    `;
    return;
  }

  trendingList.innerHTML = coins
    .map(
      (coin) => `
        <article class="trend-item">
          <span class="trend-rank">#${escapeHTML(coin.rank || "?")}</span>
          <h3>${escapeHTML(coin.name)}</h3>
          <p class="trend-symbol">${escapeHTML(coin.symbol)}</p>
          <p class="trend-price">${escapeHTML(formatPrice(coin.price))}</p>
        </article>
      `,
    )
    .join("");
}

async function loadTrending() {
  trendingState.textContent = "Loading live market movers...";

  try {
    const data = await requestJSON("/api/trending");
    renderTrending(data.coins);
    trendingState.textContent = "Updated with current trending coins.";
    lastUpdated.textContent = `Updated ${new Date().toLocaleTimeString([], {
      hour: "numeric",
      minute: "2-digit",
    })}`;
    setConnectionState("Live data connected", "state-success");
  } catch (error) {
    trendingList.innerHTML = `
      <div class="empty-state">
        <p>${escapeHTML(error.message)}</p>
      </div>
    `;
    trendingState.textContent = "Trending data is unavailable.";
    setConnectionState("Could not reach live data", "state-error");
  }
}

document.querySelectorAll("[data-fill]").forEach((button) => {
  button.addEventListener("click", () => {
    const target = document.getElementById(button.dataset.fill);
    if (!target) {
      return;
    }

    target.value = button.dataset.value || "";
    target.focus();
  });
});

singleForm.addEventListener("submit", handleSingleLookup);
compareForm.addEventListener("submit", handleComparison);
refreshTrendingButton.addEventListener("click", loadTrending);

document.querySelector("#single-symbol").value = "bitcoin";
document.querySelector("#first-symbol").value = "bitcoin";
document.querySelector("#second-symbol").value = "ethereum";

loadTrending();
handleSingleLookup(new Event("submit"));
handleComparison(new Event("submit"));
