import { useState } from 'react';
import DirectoryTree from './components/DirectoryTree';
import VideoEditor from './components/VideoEditor';

export default function App() {
  const [selectedVideoPath, setSelectedVideoPath] = useState(null);

  return (
    <div style={{ display: 'flex', height: '100vh', fontFamily: 'sans-serif' }}>
      {/* Left Sidebar: Directory Tree */}
      <aside style={{ width: '300px', borderRight: '1px solid #ccc', padding: '1rem', overflowY: 'auto' }}>
        <h3>File Explorer</h3>
        <DirectoryTree 
          path="" 
          onSelectVideo={(path) => setSelectedVideoPath(path)} 
        />
      </aside>

      {/* Right Main Content Area */}
      <main style={{ flex: 1, padding: '1rem', overflowY: 'auto' }}>
        {selectedVideoPath ? (
          <VideoEditor videoPath={selectedVideoPath} key={selectedVideoPath} />
        ) : (
          <div>
            <h3>No Video Selected</h3>
            <p>Select a video file from the directory tree on the left to start editing.</p>
          </div>
        )}
      </main>
    </div>
  );
}