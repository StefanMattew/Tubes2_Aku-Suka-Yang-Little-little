"use client";
import { motion, AnimatePresence } from "framer-motion";

export default function ResultTree({ result, elementImages }) {
  if (!result || result.length === 0) return null;

  // Kalau hasil bukan object (berarti pesan error atau array string)
  if (typeof result[0] !== "object") {
    return (
      <div className="mt-6">
        <h2 className="text-lg font-bold mb-2">Hasil:</h2>
        <ul className="bg-gray-100 p-4 rounded shadow text-sm text-gray-800 font-mono space-y-1">
          {result.map((step, index) => (
            <li key={index}>{step}</li>
          ))}
        </ul>
      </div>
    );
  }

  // Bangun tree dari akhir ke awal (reverse)
  const buildTree = (recipes, index = recipes.length - 1) => {
    if (index < 0) return null;
    const { element1, element2, result } = recipes[index];
    return {
      name: result,
      children: [
        { name: element1, children: [] },
        { name: element2, children: [] },
        buildTree(recipes, index - 1),
      ].filter(Boolean),
    };
  };

  const renderTree = (node) => {
    if (!node) return null;

    const imgSrc =
      elementImages?.[node.name] &&
      `http://localhost:8081${elementImages[node.name]}`;

    return (
      <motion.li
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className="ml-4 relative mb-2"
      >
        <div className="flex items-center space-x-2 bg-blue-100 px-3 py-1 rounded-md shadow">
          {imgSrc && (
            <img
              src={imgSrc}
              alt={node.name}
              className="w-6 h-6 rounded border border-gray-300"
            />
          )}
          <span className="text-sm font-semibold text-blue-800 font-mono">
            {node.name}
          </span>
        </div>

        {node.children?.length > 0 && (
          <motion.ul
            className="ml-4 border-l-2 border-gray-300 pl-4 mt-1"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
          >
            {node.children.map(renderTree)}
          </motion.ul>
        )}
      </motion.li>
    );
  };

  const tree = buildTree(result);

  return (
    <div className="mt-6 w-full">
      <h2 className="text-xl font-bold mb-3 text-indigo-700">
        🌳 Visualisasi Resep
      </h2>
      <ul className="pl-2">{renderTree(tree)}</ul>
    </div>
  );
}
