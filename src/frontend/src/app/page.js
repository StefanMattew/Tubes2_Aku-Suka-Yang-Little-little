"use client";
import { useEffect, useState } from "react";
import ElementCardSelector from "../components/ElementCardSelector";
import SearchButton from "../components/SearchButton";
import dynamic from "next/dynamic";
const ResultTree = dynamic(() => import("../components/ResultTree"), { ssr: false });


export default function Home() {
  const [target, setTarget] = useState("");
  const [allElements, setAllElements] = useState([]);
  const [method, setMethod] = useState("BFS");
  const [mode, setMode] = useState("single");
  const [maxRecipe, setMaxRecipe] = useState(3);
  const [result, setResult] = useState([]);
  const [elapsedTime, setElapsedTime] = useState("");
  const [visitedNodes, setVisitedNodes] = useState("");
  const [isLoading, setIsLoading] = useState(false); 

  useEffect(() => {
    setIsLoading(true);
    fetch("http://localhost:8081/elements-info")
      .then((res) => {
        if (!res.ok) throw new Error(`Gagal memuat info elemen: ${res.statusText}`);
        return res.json();
      })
      .then((allElements) => {
        setAllElements(allElements); 
      })
      .catch((error) => {
        console.error("Gagal memuat data dari backend:", error);
        setAllElements([
          { name: "Air", imagePath: "http://localhost:8081/images/air.png", tier: "Starting elements" },
          { name: "Water", imagePath: "http://localhost:8081/images/water.png", tier: "Starting elements" },
          { name: "Fire", imagePath: "http://localhost:8081/images/fire.png", tier: "Starting elements" },
          { name: "Earth", imagePath: "http://localhost:8081/images/earth.png", tier: "Starting elements" },
        ]);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);


  const handleSearch = async () => {
    if (!target) return;
    setIsLoading(true);
    setResult([]); 

    const backendURL =
      method === "BFS"
        ? "http://localhost:8081/search"
        : "http://localhost:8082/search"; 

    const body = {
      target,
      method,
      mode,
      ...(mode === "multiple" && { maxRecipe: Number(maxRecipe) }),
    };

    try {
      const res = await fetch(backendURL, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      if (!res.ok) {
        const errorData = await res.json().catch(() => ({ message: "Terjadi kesalahan jaringan atau respons tidak valid."}));
        throw new Error(errorData.message || `Gagal melakukan pencarian: ${res.statusText}`);
      }

      const data = await res.json();
      setResult(data.recipes && data.recipes.length > 0 ? data.recipes : ["Resep tidak ditemukan atau format tidak sesuai."]);
      setElapsedTime(data.elapsedTime);
      setVisitedNodes(data.visitedNodes);
    } catch (err) {
      console.error("Error ketika searching path:", err);
      setResult([`❌ ${err.message}`]);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <main className="min-h-screen bg-gradient-to-br from-indigo-100 via-purple-100 to-blue-100 flex flex-col items-center p-4 md:p-6">
      <div className="w-full max-w-5xl mx-auto">
        <h1 className="text-4xl font-bold mb-6 md:mb-10 text-center text-transparent bg-clip-text bg-gradient-to-r from-purple-600 to-indigo-600 py-2">
          Little Alchemy Solver 🧙✨
        </h1>
        {/* {element card} */}
        <ElementCardSelector
          allElements={allElements}
          selectedElement={target}
          onElementSelect={setTarget}
          />

        {/* {user's input} */}
        <div className="grid md:grid-cols-12 gap-6 md:gap-8">
          <div className="md:col-span-4 bg-white/70 backdrop-blur-md p-6 rounded-xl shadow-xl">
            <h2 className="text-xl font-semibold mb-5 text-indigo-700 border-b pb-2">Pengaturan Pencarian</h2>
            
            <div className="mb-4">
              <label className="block mb-1.5 text-gray-700 font-medium">Metode Pencarian</label>
              <select
                value={method}
                onChange={(e) => setMethod(e.target.value)}
                className="w-full p-2.5 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
                >
                <option value="BFS">BFS</option>
                <option value="DFS">DFS</option>
              </select>
            </div>

            <div className="mb-4">
              <label className="block mb-1.5 text-gray-700 font-medium">Mode Hasil</label>
              <select
                value={mode}
                onChange={(e) => setMode(e.target.value)}
                className="w-full p-2.5 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
                >
                <option value="single">Single Recipe</option>
                <option value="multiple">Multiple Recipe</option>
              </select>
            </div>

            {mode === "multiple" && (
              <div className="mb-5">
                <label className="block mb-1.5 text-gray-700 font-medium">Jumlah Maksimal Resep</label>
                <input
                  type="number"
                  min={1}
                  max={7} 
                  value={maxRecipe}
                  onChange={(e) => setMaxRecipe(Number(e.target.value))}
                  className="w-full p-2.5 border border-gray-300 rounded-md shadow-sm focus:ring-indigo-500 focus:border-indigo-500"
                  />
              </div>
            )}

            <SearchButton
              label={isLoading ? "Mencari..." : `Cari Resep (${method})`}
              onClick={handleSearch}
              disabled={!target || isLoading}
              />
            {isLoading && (
              <div className="flex justify-center items-center mt-4">
                <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-indigo-600"></div>
                <p className="ml-2 text-sm text-indigo-600">Sedang memproses...</p>
              </div>
            )}
          </div>

          {/* {tree result} */}
          <div className="md:col-span-8 bg-white/70 backdrop-blur-md p-6 rounded-xl shadow-xl min-h-[300px]">
            {result && result.length > 0 ? (
              <ResultTree
                targetElement={target}
                recipeSteps={result}
                time={elapsedTime}
                nodes={visitedNodes}
                elementImages={Object.fromEntries(allElements.map(el => [el.name, el.imagePath]))}
                live={true}
              />
              // <ResultTree targetElement={target} recipeSteps={result} elementImages={elementImages} />
            ) : (
              <div className="flex flex-col items-center justify-center h-full text-center">
                {!isLoading && !target && <p className="text-gray-500 text-lg">Pilih elemen target untuk memulai.</p>}
                {!isLoading && target && result.length === 0 && <p className="text-gray-500 text-lg">Belum ada hasil. Klik "Cari Resep" untuk melihat visualisasi.</p>}
              </div>
            )}
          </div>
        </div>
      </div>
    </main>
  );
}