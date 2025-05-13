"use client";
import { Tree, TreeNode } from 'react-organizational-chart';
import styled from 'styled-components';

const StyledNode = styled.div`
  padding: 8px 12px;
  background-color: ${(props) =>
    props.$isRoot ? '#6366f1' : props.$isLeaf ? '#bbf7d0' : '#ddd6fe'};
  color: ${(props) => (props.$isRoot ? '#ffffff' : '#1f2937')};
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  font-family: monospace;
  font-size: 0.9rem;
  box-shadow: 0 2px 6px rgba(0,0,0,0.15);
  gap: 8px;
  transition: transform 0.2s ease;

  &:hover {
    transform: scale(1.03);
  }
`;

export default function ResultTree({ targetElement, recipeSteps, time, nodes, elementImages }) {
  if (!recipeSteps || recipeSteps.length === 0 || typeof recipeSteps[0] === "string") {
    return (
      <div className="mt-4 text-sm text-gray-700">
        <p>{targetElement ? `Hasil untuk ${targetElement}:` : 'Hasil:'}</p>
        <ul className="mt-2 bg-yellow-100 p-3 rounded-md">
          {recipeSteps.map((step, index) => <li key={index}>{step}</li>)}
        </ul>
      </div>
    );
  }

  const actualRecipePath = Array.isArray(recipeSteps[0]) ? recipeSteps[0] : recipeSteps;

  const buildTreeFromSteps = (target, steps) => {
    const map = new Map();
    steps.forEach(step => map.set(step.result, step));

    const buildNode = (element) => {
      const step = map.get(element);
      if (!step) return { name: element, children: [] };
      return {
        name: element,
        children: [buildNode(step.element1), buildNode(step.element2)],
      };
    };

    return buildNode(target);
  };

  const treeData = buildTreeFromSteps(targetElement, actualRecipePath);

  const renderTree = (node, isRoot = false, key = null) => {
    if (!node || !node.name) return null; // ⛑️ Penjagaan penting

    const isLeaf = !node.children || node.children.length === 0;
    const imageSrc = elementImages[node.name] || "/images/elements/placeholder.png";

    return (
      <TreeNode
        key={key || node.name}
        label={
          <StyledNode $isRoot={isRoot} $isLeaf={isLeaf}>
            <img
              src={imageSrc}
              alt={node.name}
              width={28}
              height={28}
              style={{ borderRadius: 4, border: '1px solid #e5e7eb' }}
              onError={(e) => { e.target.src = "/images/elements/placeholder.png"; }}
            />
            {node.name}
          </StyledNode>
        }
      >
        {node.children?.map((child, idx) =>
          child ? renderTree(child, false, `${node.name}-${idx}`) : null
        )}
      </TreeNode>
    );
  };

  return (
    <div className=" flex flex-col overflow-hidden">
      {/* Header */}
      <div className="mb-2">
        <h2 className="text-xl font-bold text-indigo-700">
          Recipe Tree: <span className="text-purple-600">{targetElement}</span>
        </h2>
      </div>

      {/* Tree */}
      <div className="flex-1 overflow-auto border rounded p-2">
        <Tree
          lineWidth={"2px"}
          lineColor={"#9ca3af"}
          lineBorderRadius={"10px"}
          label={<StyledNode $isRoot>{targetElement}</StyledNode>}
        >
          {treeData.children?.map((child, idx) =>
            child ? renderTree(child, false, `${treeData.name}-${idx}`) : null
          )}
        </Tree>
      </div>

      {/* Searching info */}
      <div className="grid grid-cols-2 gap-3 mt-4">
        <h2 className="text-sm font-semibold text-indigo-700">
          Elapsed Time: <span className="text-purple-600">{time}</span>
        </h2>
        <h2 className="text-sm font-semibold text-indigo-700">
          Visited Node: <span className="text-purple-600">{nodes}</span>
        </h2>
      </div>
    </div>
  );
}
