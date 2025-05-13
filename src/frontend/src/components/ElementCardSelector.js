import React, { useState } from "react";

export default function ElementCardSelector({ allElements, selectedElement, onElementSelect }) {
  const [selectedTier, setSelectedTier] = useState("Starting elements");

  if (!allElements || allElements.length === 0) return <p className="text-gray-500">Memuat elemen...</p>;

  // Ambil semua tier unik & sortir
  const tierOptions = Array.from(new Set(allElements.map((el) => el.tier))).sort((a, b) => {
    if (a.includes("Starting")) return -1;
    if (b.includes("Starting")) return 1;
    const aNum = parseInt(a.match(/\d+/));
    const bNum = parseInt(b.match(/\d+/));
    return (aNum || 100) - (bNum || 100);
  });

  // Filter elemen berdasarkan tier yang dipilih
  const filteredElements = allElements.filter((el) => el.tier === selectedTier);

  return (
    <div className="mb-8 p-4 bg-white/70 backdrop-blur-md rounded-lg shadow-lg">
      <div className="grid grid-cols-2 gap-3 mb-4">
        <h2 className="text-2xl font-semibold text-indigo-700">Pilih Elemen Target</h2>
        <select
          className="border px-3 py-2 rounded-md shadow-sm focus:outline-none focus:ring focus:border-indigo-500 ml-auto"
          value={selectedTier}
          onChange={(e) => setSelectedTier(e.target.value)}
        >
          {tierOptions.map((tier) => (
            <option key={tier} value={tier}>
              {tier}
            </option>
          ))}
        </select>
      </div>
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 lg:grid-cols-6 gap-3">
        {filteredElements.map((el) => (
          <button
            key={`${el.tier}-${el.name}`}
            onClick={() => onElementSelect(el.name)}
            className={`flex flex-col items-center p-2 rounded-lg transition hover:scale-105 shadow-sm ${
              selectedElement === el.name
                ? "bg-indigo-500 text-white ring-2 ring-indigo-700"
                : "bg-white hover:bg-indigo-100"
            }`}
          >
            <img
              src={el.imagePath}
              alt={el.name}
              className="w-12 h-12 object-contain mb-1"
              onError={(e) => (e.target.src = "/images/elements/placeholder.png")}
            />
            <span className="text-xs text-center font-medium">{el.name}</span>
          </button>
        ))}
      </div>
    </div>
  );
}
