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
  
  const actualRecipePath = (Array.isArray(recipeSteps) && recipeSteps.length > 0 && Array.isArray(recipeSteps[0])) 
                           ? recipeSteps[0] 
                           : recipeSteps;   

  const buildTreeFromSteps = (target, stepsToProcess) => {

    if (!stepsToProcess || !Array.isArray(stepsToProcess) || stepsToProcess.length === 0) {
      console.warn("buildTreeFromSteps - target:", target, ". Mengembalikkan leaf node.");
      return { name: target, children: [] };
    }

    const reversedSteps = [...stepsToProcess].reverse(); 
    
    const finalStep = reversedSteps.find(step => step && typeof step.result !== 'undefined' && step.result === target);

    if (!finalStep) {
      console.warn("buildTreeFromSteps - Tidak ada recipe yang ditemukan secara langsung untuk main target:", target, ".");
      const isBaseElementInSteps = stepsToProcess.some(step => step && (step.element1 === target || step.element2 === target));
      if (!stepsToProcess.some(step => step && step.result === target) && isBaseElementInSteps) {
         console.log("buildTreeFromSteps - Target", target, ".");
         return { name: target, children: [] };
      }
      return { name: target, children: [] };
    }

    const buildNode = (currentResultElement) => {

      if (!currentResultElement || typeof currentResultElement !== 'string' || currentResultElement.trim() === "") {
          console.warn("buildNode - Invalid currentResultElement:", currentResultElement, ". Returning null.");
          return null; 
      }

      const producingStep = reversedSteps.find(s => s && typeof s.result === 'string' && s.result === currentResultElement);

      if (!producingStep) {
        return { name: currentResultElement, children: [] };
      }

      if (typeof producingStep.element1 !== 'string' || typeof producingStep.element2 !== 'string') {
          console.warn("buildNode - currentResultElement", currentResultElement, "invalid element1 or element2:", JSON.parse(JSON.stringify(producingStep)));
          return { name: currentResultElement, children: [] }; // Treat as leaf if ingredients are malformed
      }

      const child1Node = buildNode(producingStep.element1);
      const child2Node = buildNode(producingStep.element2);
      
      const children = [child1Node, child2Node].filter(Boolean);
      
      return {
        name: currentResultElement,
        children: children 
      };
    };

    const rootNode = buildNode(target);
    return rootNode;
  };

  const treeData = buildTreeFromSteps(targetElement, actualRecipePath);
  console.log("ResultTree - treeData:", treeData ? JSON.parse(JSON.stringify(treeData)) : "null or undefined");


  const renderTreeNode = (node, depth = 0, index = 0, isRoot = false) => {
    if (!node || typeof node.name === 'undefined') { 
        return null;
    }
    const imgSrc = elementImages[node.name];
    const hasChildren = node.children && node.children.length > 0;

    let nodeBgColor = 'bg-purple-100';
    let nodeTextColor = 'text-purple-800';
    let nodePStyle = 'px-3 py-1.5 rounded-lg shadow-md';

    if (isRoot) {
      nodeBgColor = 'bg-indigo-500';
      nodeTextColor = 'text-white';
      nodePStyle = 'px-4 py-2 rounded-lg shadow-lg text-base';
    } else if (!hasChildren) { 
      nodeBgColor = 'bg-green-100';
      nodeTextColor = 'text-green-800';
    }

    return (
      <motion.li
        key={`${node.name}-${depth}-${index}`} 
        initial={{ opacity: 0, x: -20 }}
        animate={{ opacity: 1, x: 0 }}
        transition={{ delay: depth * 0.2 + index * 0.05, duration: 0.3 }}
        className={`relative ${hasChildren ? 'pl-8' : 'pl-8'} mb-2 list-none`}
      >
        {!isRoot && (
          <div className="absolute left-3 top-0 bottom-1/2 w-px bg-slate-400"></div>
        )}
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
          <ul className="mt-2"> 
            {node.children.map((child, idx) => (
              <Fragment key={`${node.name}-child-${child && child.name ? child.name : `unknownchild-${idx}`}-${idx}`}> {/* Added safety for child.name */}
                {renderTreeNode(child, depth + 1, idx, false)}
              </Fragment>
            ))}
          </ul>
        )}
      </motion.li>
    );
  };
  
  if (!treeData) {
    console.error("ResultTree - treeData = null atau undefined.");
    return <p className="text-red-500">Error: Tidak dapat membuat data pohon.</p>;
  }

  return (
    <div className="mt-2 w-full overflow-x-auto custom-scrollbar">
      <h2 className="text-2xl font-bold mb-4 text-indigo-700">
        Recipe Tree: <span className="text-purple-600">{targetElement}</span>
      </h2>
      {treeData && treeData.name ? ( 
        <ul className="pl-2"> 
          {renderTreeNode(treeData, 0, 0, true)}
        </ul>
      ) : (
        <p className="text-gray-600">Tidak dapat membuat pohon visualisasi.</p>
      )}
    </div>
  );
}
