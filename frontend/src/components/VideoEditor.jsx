import { useRef, useState } from "react";

export default function VideoEditor({ videoPath }) {
  const videoRef = useRef(null);
  const [segments, setSegments] = useState([]);
  const [currentStart, setCurrentStart] = useState(null);

  const handleSetPoint = () => {
    const currentTime = Math.floor(videoRef.current.currentTime);

    if (currentStart === null) {
      setCurrentStart(currentTime);
    } else {
      setSegments([...segments, { start: currentStart, end: currentTime }]);
      setCurrentStart(null);
    }
  };

  const handleSendForProcessing = async () => {
    const payload = {
      video_path: videoPath,
      segments: segments,
    };

    const res = await fetch("/api/process", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    if (res.ok) alert("Processing started!");
  };

  return (
    <div>
      <h2>Editing: {videoPath}</h2>

      <video ref={videoRef} controls width="100%" style={{ maxWidth: "800px" }}>
        <source src={`/media/${videoPath}`} type="video/mp4" />
      </video>

      <div style={{ marginTop: "1rem" }}>
        <button onClick={handleSetPoint}>
          {currentStart === null
            ? "Set Segment Start"
            : `Set Segment End (Start: ${currentStart}s)`}
        </button>
      </div>

      <h4 style={{ marginTop: "1.5rem" }}>Configured Segments:</h4>
      <ul>
        {segments.map((s, idx) => (
          <li key={idx}>
            Start: {s.start}s — End: {s.end}s
          </li>
        ))}
      </ul>

      <button
        onClick={handleSendForProcessing}
        disabled={segments.length === 0}
        style={{ marginTop: "1rem", padding: "0.5rem 1rem" }}
      >
        Send for processing
      </button>
    </div>
  );
}
