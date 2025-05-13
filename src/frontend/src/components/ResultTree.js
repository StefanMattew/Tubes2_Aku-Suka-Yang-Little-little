"use client"
import { motion } from 'framer-motion';
import { Fragment } from 'react';

export default function ResultTree({ targetElement, recipeSteps, elementImages }) {
  if (!recipeSteps || recipeSteps.length === 0 || (recipeSteps.length > 0 && typeof recipeSteps[0] === 'string')) {
    if (!recipeSteps || recipeSteps.length === 0) return null; 
    return (
      <div className="mt-6">
        <h2 className="text-xl font-bold mb-3 text-indigo-700">
          {targetElement ? `Hasil untuk ${targetElement}:` : 'Hasil:'}
        </h2>
        <div className="bg-yellow-50 border border-yellow-300 text-yellow-800 p-4 rounded-md shadow text-sm font-mono space-y-1">
          {recipeSteps.map((step, index) => (
            <div key={index}>{step}</div>
          ))}
        </div>
      </div>
    );
  }
  
  const actualRecipePath = (Array.isArray(recipeSteps) && recipeSteps.length > 0 && Array.isArray(recipeSteps[0])) ? recipeSteps[0] : recipeSteps;   

  const buildTreeFromSteps = (target, stepsToProcess) => {
    console.log("Build tree for:", target);
    console.log("Steps:", stepsToProcess);
    if (!stepsToProcess || stepsToProcess.length === 0) return { name: target, children: [] };

    if (!stepsToProcess || !Array.isArray(stepsToProcess) || stepsToProcess.length === 0) {
      console.warn("Tidak ada langkah-langkah valid untuk diproses dalam buildTreeFromSteps setelah ekstraksi.");
      return { name: target, children: [] };
    }

    const reversedSteps = [...stepsToProcess].reverse(); 
    
    let rootNode = { name: target, children: [] };
    
    const finalStep = reversedSteps.find(step => step && typeof step.result !== 'undefined' && step.result === target);

    if (!finalStep) {
      console.warn("Tidak ada langkah resep yang menghasilkan target utama:", target, reversedSteps);
      const isBaseElementInSteps = stepsToProcess.some(step => step.element1 === target || step.element2 === target);

      if (!stepsToProcess.some(step => step.result === target) && isBaseElementInSteps) {
        return { name: target, children: [] }; // target adalah salah satu elemen dasar dalam resep yang diberikan tapi bukan hasil
      }
      return { name: target, children: [] };
    }

    // Fungsi rekursif untuk membangun node dan anak-anaknya
    // `currentResultElement` adalah elemen yang sedang kita cari resepnya
    const buildNode = (currentResultElement) => {
        const producingStep = reversedSteps.find(s => s && typeof s.result !== 'undefined' && s.result === currentResultElement);
        if (!producingStep) {
            return { name: currentResultElement, children: [] };
        }

        const child1 = buildNode(producingStep.element1);
        const child2 = buildNode(producingStep.element2);
        
        return {
            name: currentResultElement,
            children: [child1, child2].filter(Boolean) 
        };
    };

    rootNode = buildNode(target);
    
    return rootNode;
  };


  const renderTreeNode = (node, depth = 0, index = 0, isRoot = false) => {
    if (!node) return null;
    const imgSrc = elementImages[node.name];
    const hasChildren = node.children && node.children.length > 0;

    // Styling untuk node
    let nodeBgColor = 'bg-purple-100';
    let nodeTextColor = 'text-purple-800';
    let nodePStyle = 'px-3 py-1.5 rounded-lg shadow-md';

    if (isRoot) {
      nodeBgColor = 'bg-indigo-500';
      nodeTextColor = 'text-white';
      nodePStyle = 'px-4 py-2 rounded-lg shadow-lg text-base';
    } else if (!hasChildren) { // Leaf node (elemen dasar dari resep)
      nodeBgColor = 'bg-green-100';
      nodeTextColor = 'text-green-800';
    }


    return (
      <motion.li
        key={`${node.name}-${depth}-${index}`}
        initial={{ opacity: 0, x: -20 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ delay: depth * 0.2 + index * 0.05, duration: 0.3 }}
        className={`relative ${hasChildren ? 'pl-8' : 'pl-8'} mb-2 list-none`} // pl-8 untuk memberi ruang bagi garis
      >
        {/* Garis vertikal dari parent (kecuali root) */}
        {!isRoot && (
          <div className="absolute left-3 top-0 bottom-1/2 w-px bg-slate-400"></div>
        )}
        {/* Garis horizontal ke node */}
        {!isRoot && (
           <div className="absolute left-3 top-1/2 h-px w-5 bg-slate-400 transform -translate-y-1/2"></div>
        )}

        <div className={`flex items-center space-x-2 ${nodeBgColor} ${nodePStyle} inline-block`}>
          {imgSrc && (
            <img
              src={imgSrc}
              alt={node.name}
              className={`w-7 h-7 rounded border ${isRoot ? 'border-indigo-300' : 'border-purple-300'}`}
              onError={(e) => { e.target.onerror = null; e.target.src = '/images/elements/placeholder.png'; }}
            />
          )}
          <span className={`font-semibold ${nodeTextColor} font-mono ${isRoot ? 'text-background' : 'text-sm'}`}>
            {node.name}
          </span>
        </div>

        {hasChildren && (
          <ul className="mt-2"> {/* Tidak perlu border kiri di sini, diatur oleh pseudo-elemen atau garis absolut */}
            {node.children.map((child, idx) => (
              <Fragment key={`${node.name}-child-${child.name}-${idx}`}>
                {renderTreeNode(child, depth + 1, idx, false)}
              </Fragment>
            ))}
          </ul>
        )}
      </motion.li>
    );
  };
  
  // Bangun pohon dari target dan langkah-langkah resep
  const treeData = buildTreeFromSteps(targetElement, actualRecipePath);

  return (
    <div className="mt-2 w-full overflow-x-auto custom-scrollbar"> {/* overflow-x-auto di sini */}
      <h2 className="text-2xl font-bold mb-4 text-indigo-700">
        Recipe Tree: <span className="text-purple-600">{targetElement}</span>
      </h2>
      {treeData ? (
        <ul className="pl-2"> {/* Hapus padding kiri jika garis diatur oleh li */}
          {renderTreeNode(treeData, 0, 0, true)}
        </ul>
      ) : (
        <p className="text-gray-600">Tidak dapat membuat pohon visualisasi.</p>
      )}
    </div>
  );
}