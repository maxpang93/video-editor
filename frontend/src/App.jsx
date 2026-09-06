import { useState, useRef } from "react";
import DirectoryTree from "./components/DirectoryTree";
import VideoEditor from "./components/VideoEditor";

export default function App() {
  const [selectedVideoPath, setSelectedVideoPath] = useState(null);
  const [sidebarWidth, setSidebarWidth] = useState(300);
  const isResizing = useRef(false);

  const handleMouseDown = (e) => {
    e.preventDefault();
    isResizing.current = true;

    document.addEventListener("mousemove", handleMouseMove);
    document.addEventListener("mouseup", handleMouseUp);
  };

  const handleMouseMove = (e) => {
    if (!isResizing.current) return;

    const newWidth = Math.min(Math.max(e.clientX, 150), 600);
    setSidebarWidth(newWidth);
  };

  const handleMouseUp = () => {
    isResizing.current = false;
    document.removeEventListener("mousemove", handleMouseMove);
    document.removeEventListener("mouseup", handleMouseUp);
  };

  return (
    <div
      style={{
        display: "flex",
        height: "100%",
        width: "100%",
        overflow: "hidden",
        fontFamily: "sans-serif",
      }}
    >
      {/* Left Sidebar Pane */}
      <aside
        style={{
          width: `${sidebarWidth}px`,
          height: "100%",
          flexShrink: 0,
          padding: "1rem",
          overflowY: "auto",
          boxSizing: "border-box",
        }}
      >
        <h3>File Explorer</h3>
        <DirectoryTree
          path=""
          onSelectVideo={(path) => setSelectedVideoPath(path)}
        />
      </aside>

      {/* Draggable Divider */}
      <div
        onMouseDown={handleMouseDown}
        style={{
          width: "5px",
          flexShrink: 0, // Prevents divider from squishing
          cursor: "col-resize",
          backgroundColor: "#ccc",
          userSelect: "none",
          transition: "background-color 0.2s",
        }}
        onMouseEnter={(e) => (e.target.style.backgroundColor = "#999")}
        onMouseLeave={(e) => (e.target.style.backgroundColor = "#ccc")}
      />

      {/* Right Main Pane */}
      <main
        style={{
          flex: 1,
          height: "100%",
          minWidth: 0, // CRITICAL: Allows Flexbox item to shrink past content min-size
          padding: "1rem",
          overflowY: "auto",
          overflowX: "hidden", // Forces wide content inside main to truncate or stay contained
          boxSizing: "border-box",
        }}
      >
        {selectedVideoPath ? (
          <VideoEditor videoPath={selectedVideoPath} key={selectedVideoPath} />
        ) : (
          <div>
            <h3>No Video Selected</h3>
            <p>
              Select a video file from the directory tree on the left to start
              editing.
            </p>
          </div>
        )}
      </main>
    </div>
  );
}
