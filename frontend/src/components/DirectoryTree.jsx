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
    <ul style={{ listStyleType: "none", paddingLeft: path ? "1rem" : "0" }}>
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
          <li key={itemPath} style={{ margin: "4px 0" }}>
            <span
              onClick={() => onSelectVideo(itemPath)}
              style={{ cursor: "pointer", color: "#0066cc" }}
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
      <span
        onClick={() => setExpanded(!expanded)}
        style={{ cursor: "pointer", fontWeight: "bold", userSelect: "none" }}
      >
        {expanded ? "📂" : "📁"} {item.name}
      </span>

      {/* Render child tree lazy-loaded only when expanded */}
      {expanded && <DirectoryTree path={path} onSelectVideo={onSelectVideo} />}
    </li>
  );
}
