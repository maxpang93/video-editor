import { useState, useEffect } from "react";

export default function DirectoryTree({ path = "", onSelectVideo }) {
  const [items, setItems] = useState([]);

  // Fetch directory contents for this path level
  useEffect(() => {
    fetch(`/api/files?path=${encodeURIComponent(path)}`)
      .then((res) => res.json())
      .then((data) => setItems(data))
      .catch(console.error);
  }, [path]);

  return (
    <ul
      style={{
        listStyleType: "none",
        textAlign: "left", // Ensures left alignment regardless of global/parent CSS
        paddingLeft: path ? "1.25rem" : "0", // Adds deeper indent for sub-trees
        margin: 0,
        borderLeft: path ? "2px solid #ccc" : "none", // Adds a subtle border guide for nested trees
      }}
    >
      {items.map((item) => {
        const itemPath = path ? `${path}/${item.name}` : item.name;

        if (item.isFolder) {
          return (
            <FolderItem
              key={itemPath}
              item={item}
              path={itemPath}
              onSelectVideo={onSelectVideo}
            />
          );
        }

        return (
          <li key={itemPath} style={{ margin: "4px 0", paddingLeft: "0.5rem" }}>
            <span
              onClick={() => onSelectVideo(itemPath)}
              style={{
                cursor: "pointer",
                color: "#0066cc",
                display: "inline-block",
              }}
            >
              🎥 {item.name}
            </span>
          </li>
        );
      })}
    </ul>
  );
}

// Helper component that manages its own expanded/collapsed state
function FolderItem({ item, path, onSelectVideo }) {
  const [expanded, setExpanded] = useState(false);

  return (
    <li style={{ margin: "4px 0" }}>
      <div
        onClick={() => setExpanded(!expanded)}
        style={{
          cursor: "pointer",
          fontWeight: "bold",
          userSelect: "none",
          display: "flex",
          alignItems: "center",
          gap: "6px",
        }}
      >
        <span>{expanded ? "📂" : "📁"}</span>
        <span>{item.name}</span>
      </div>
      {/* Render child tree lazy-loaded only when expanded */}
      {expanded && <DirectoryTree path={path} onSelectVideo={onSelectVideo} />}
    </li>
  );
}
