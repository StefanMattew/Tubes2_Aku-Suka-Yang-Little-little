"use client";
import { useEffect, useState } from "react";
import ElementSelector from "../components/ElementSelector";
import SearchButton from "../components/SearchButton";
import ResultTree from "../components/ResultTree";

export default function Home() {
  const [target, setTarget] = useState("");
  const [elements, setElements] = useState([]);
  const [method, setMethod] = useState("BFS");
  const [mode, setMode] = useState("single");
  const [maxRecipe, setMaxRecipe] = useState(3);
  const [result, setResult] = useState([]);
  const [elementImages, setElementImages] = useState({});


useEffect(() => {
  fetch("http://localhost:8081/elements") // target
    .then((res) => res.json())
    .then((data) => setElements(data.elements))
    .catch(() => {
      console.error("Gagal memuat elemen.");
      setElements(["Air", "Water", "Fire", "Earth"]);
    });

  fetch("http://localhost:8081/element-images") // images for visualizing
    .then((res) => res.json())
    .then((data) => setElementImages(data))
    .catch(() => {
      console.error("Gagal memuat gambar elemen.");
    });
}, []);

  const handleSearch = async () => {
    const backendURL =
      method === "BFS"
        ? "http://localhost:8081/search"
        : "http://localhost:8082/search";

    const body = {
      target,
      method,
      mode,
      maxRecipe,
    };

    try {
      const res = await fetch(backendURL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      const data = await res.json();
      const steps = data.recipes?.[0] || ["❌ Tidak ditemukan."];

      setResult([
        ...steps,
        `⏱ Waktu: ${data.elapsedTime}ms`,
        `🧠 Node dikunjungi: ${data.visitedNodes}`,
      ]);
    } catch (err) {
      console.error(err);
      setResult(["❌ Terjadi kesalahan saat fetch."]);
    }
  };

  return (
<main className="min-h-screen bg-gradient-to-br from-indigo-100 via-blue-100 to-white flex items-center justify-center p-6">
  <div className="w-full max-w-md bg-white/80 backdrop-blur-md p-6 rounded-xl shadow-2xl">
    <h1 className="text-3xl font-bold mb-6 text-center text-indigo-700">Little Alchemy Solver</h1>
        <ElementSelector
          label="Target Elemen"
          value={target}
          onChange={setTarget}
          options={elements}
        />

        <div className="mb-4">
          <label className="block mb-1 text-gray-700 font-semibold">Metode Pencarian</label>
          <select
            value={method}
            onChange={(e) => setMethod(e.target.value)}
            className="w-full p-2 border border-gray-300 rounded"
          >
            <option value="BFS">BFS</option>
            <option value="DFS">DFS</option>
          </select>
        </div>

        <div className="mb-4">
          <label className="block mb-1 text-gray-700 font-semibold">Mode</label>
          <select
            value={mode}
            onChange={(e) => setMode(e.target.value)}
            className="w-full p-2 border border-gray-300 rounded"
          >
            <option value="single">Satu Resep Terpendek</option>
            <option value="multiple">Multiple Resep</option>
          </select>
        </div>

        {mode === "multiple" && (
          <div className="mb-4">
            <label className="block mb-1 text-gray-700 font-semibold">Jumlah Maksimal Resep</label>
            <input
              type="number"
              min={1}
              value={maxRecipe}
              onChange={(e) => setMaxRecipe(Number(e.target.value))}
              className="w-full p-2 border border-gray-300 rounded"
            />
          </div>
        )}

        <SearchButton
          label={`Cari dengan ${method}`}
          onClick={handleSearch}
          disabled={!target}
        />
        {/* <button
          onClick={() => setResult([])}
          className="mt-4 px-3 py-1 bg-red-200 rounded text-sm"
        >
          🔄 Reset Tree
        </button> */}
        <ResultTree result={result} elementImages={elementImages} />
      </div>
    </main>
  );
}
