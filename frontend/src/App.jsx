import { useState } from "react";
import "./App.css";

const API_URL = "http://localhost:8080/api/restaurants";

function App() {
  const [query, setQuery] = useState("");
  const [restaurants, setRestaurants] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [searched, setSearched] = useState(false);

  async function searchRestaurants(searchQuery = "") {
    setLoading(true);
    setError("");

    try {
      const params = new URLSearchParams();
      const trimmedQuery = searchQuery.trim();

      if (trimmedQuery) {
        params.set("q", trimmedQuery);
      }

      const queryString = params.toString();
      const url = queryString
        ? `${API_URL}?${queryString}`
        : API_URL;

      const response = await fetch(url);
      const result = await response.json();

      if (!response.ok || !result.success) {
        throw new Error(
          result.message || "Could not load restaurants. Please try again."
        );
      }

      setRestaurants(Array.isArray(result.data) ? result.data : []);
      setSearched(true);
    } catch (err) {
      setRestaurants([]);
      setSearched(true);

      setError(
        err.message === "Failed to fetch"
          ? "Could not connect to the server. Check that the backend is running."
          : err.message || "Something went wrong. Please try again."
      );
    } finally {
      setLoading(false);
    }
  }

  function handleSubmit(event) {
    event.preventDefault();
    searchRestaurants(query);
  }

  return (
    <main className="page">
      <section className="search-panel">
        <p className="eyebrow">FIND YOUR NEXT MEAL</p>

        <h1>Find restaurants</h1>

        <p className="subtitle">
          Search by restaurant name, city, or address.
        </p>

        <form className="search-form" onSubmit={handleSubmit}>
          <input
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Try pizza or Bristol..."
            aria-label="Search restaurants by name or location"
          />

          <button type="submit" disabled={loading}>
            {loading ? "Searching..." : "Search"}
          </button>
        </form>

        {loading && (
          <p className="loading-message" role="status">
            Loading restaurants...
          </p>
        )}

        {!loading && error && (
          <div className="error-message" role="alert">
            <p>{error}</p>
            <button
              type="button"
              onClick={() => searchRestaurants(query)}
            >
              Try again
            </button>
          </div>
        )}

        {!loading && !error && searched && restaurants.length === 0 && (
          <div className="empty-state" role="status">
            <h2>No restaurants found</h2>
            <p>Try another restaurant name or location.</p>
          </div>
        )}

        {!loading && !error && restaurants.length > 0 && (
          <section className="results" aria-live="polite">
            <h2>Restaurants ({restaurants.length})</h2>

            <div className="restaurant-list">
              {restaurants.map((restaurant) => (
                <article className="restaurant-card" key={restaurant.id}>
                  <h3>{restaurant.name}</h3>

                  {restaurant.description && (
                    <p>{restaurant.description}</p>
                  )}

                  <p className="location">
                    {[restaurant.address, restaurant.city, restaurant.state]
                      .filter(Boolean)
                      .join(", ")}
                  </p>

                  {restaurant.opening_time && restaurant.closing_time && (
                    <p className="hours">
                      {restaurant.opening_time} – {restaurant.closing_time}
                    </p>
                  )}
                </article>
              ))}
            </div>
          </section>
        )}
      </section>
    </main>
  );
}

export default App;
