import { BrowserRouter, Route, Routes } from "react-router-dom";
import "./App.css";
import ChatScreen from "./components/Chat-room";
import Home from "./components/Home";

function App() {
  return (
    <div>
      <BrowserRouter>
        <Routes>
          <Route path="/chat-room" element={<ChatScreen />} />
          <Route path="/" element={<Home />} />
        </Routes>
      </BrowserRouter>
    </div>
  );
}

export default App;
