"use client";

import { useEffect, useState } from "react";
import Image from "next/image";

type Offer = {
  id: string;
  title: string;
  description: string;
  payout: number;
  vendor_id: string;
  category: string;
};

type StorefrontData = {
  offers: Offer[];
  user: string;
};

const CATEGORIES = ["all", "Electronics", "Fitness", "Fashion", "Home", "Health", "Learning"];

export default function StorefrontHome() {
  const [data, setData] = useState<StorefrontData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [activeCategory, setActiveCategory] = useState("all");

  useEffect(() => {
    // In production cluster, this would hit the API gateway or BFF internal service
    // For local dev, adjust as needed. The backend JSON BFF is expected to serve at the root '/'
    const BFF_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3000";

    async function fetchOffers() {
      try {
        const res = await fetch(`${BFF_URL}/`, {
          headers: {
            "Accept": "application/json",
          },
        });
        if (!res.ok) throw new Error("Failed to fetch products");
        const json = await res.json();
        setData(json);
      } catch (err: any) {
        setError(err.message || "An error occurred");
      } finally {
        setLoading(false);
      }
    }

    fetchOffers();
  }, []);

  const handleAddToCart = async (offerId: string) => {
    // Fire to the BFF or API gateway
    console.log("Adding to cart:", offerId);
    alert("Added to cart! (Wire this to the Zustand store/backend)");
  };

  if (loading) {
    return (
      <div className="container py-16 text-center">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-gray-200 rounded w-1/4 mx-auto"></div>
          <div className="h-4 bg-gray-200 rounded w-1/2 mx-auto"></div>
        </div>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="container py-16 text-center text-red-500">
        <h2 className="text-xl font-bold mb-2">Error Loading Storefront</h2>
        <p>{error}</p>
      </div>
    );
  }

  const filteredOffers =
    activeCategory === "all"
      ? data.offers || []
      : (data.offers || []).filter((o) => o.category === activeCategory);

  return (
    <>
      {/* Hero Section */}
      <div className="bg-blue-600 text-white py-16 mb-8">
        <div className="container text-center">
          <h1 className="text-4xl font-extrabold tracking-tight sm:text-5xl lg:text-6xl mb-4">
            Shop by Category
          </h1>
          <p className="mt-4 max-w-xl mx-auto text-xl text-blue-100">
            Microservice Enterprise E-Commerce
          </p>
        </div>
      </div>

      {/* Category Tabs */}
      <div className="container mb-8">
        <div className="flex flex-wrap gap-2 justify-center">
          {CATEGORIES.map((cat) => (
            <button
              key={cat}
              onClick={() => setActiveCategory(cat)}
              className={`px-4 py-2 rounded-full text-sm font-medium transition-colors ${activeCategory === cat
                  ? "bg-blue-600 text-white shadow-md"
                  : "bg-white text-gray-700 hover:bg-gray-100 border border-gray-200"
                }`}
            >
              {cat === "all" ? "All Products" : cat}
            </button>
          ))}
        </div>
      </div>

      {/* Product Grid */}
      <div className="container">
        {filteredOffers.length === 0 ? (
          <div className="text-center py-12 text-gray-500 bg-white rounded-lg border border-dashed">
            No products found for this category.
          </div>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
            {filteredOffers.map((offer) => (
              <div
                key={offer.id}
                className="group relative flex flex-col overflow-hidden rounded-lg border bg-white shadow-sm transition-all hover:shadow-md"
              >
                <div className="aspect-square bg-gray-200 relative overflow-hidden">
                  <div className="absolute inset-0 flex items-center justify-center text-gray-400">
                    {/* Placeholder image representation */}
                    <svg className="h-12 w-12 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                  </div>
                </div>

                <div className="flex flex-1 flex-col p-4">
                  <div className="mb-2 flex items-center justify-between">
                    <span className="inline-flex items-center rounded-md bg-blue-50 px-2 py-1 text-xs font-medium text-blue-700 ring-1 ring-inset ring-blue-700/10">
                      {offer.category}
                    </span>
                    <span className="text-lg font-bold text-gray-900">${offer.payout.toFixed(2)}</span>
                  </div>

                  <h3 className="text-base font-semibold text-gray-900 mb-1 line-clamp-1">
                    {offer.title}
                  </h3>
                  <p className="text-sm text-gray-500 line-clamp-2 mb-4 flex-1">
                    {offer.description}
                  </p>

                  <button
                    onClick={() => handleAddToCart(offer.id)}
                    className="mt-auto w-full rounded-md bg-blue-600 px-3 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-500 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 transition-all"
                  >
                    Add to cart
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </>
  );
}
