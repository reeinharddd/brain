import { useEffect, useState, useRef } from "react";

interface Status {
  status: string;
  time: string;
  processes: number;
}

interface Processes {
  [key: string]: string;
}

export default function App() {
  const [status, setStatus] = useState<Status | null>(null);
  const [processes, setProcesses] = useState<Processes>({});
  const [logs, setLogs] = useState<string[]>([]);
  const logsEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [logs]);

  const fetchData = async () => {
    try {
      const stRes = await fetch("http://localhost:9090/api/status");
      const stData = await stRes.json();
      setStatus(stData);

      const psRes = await fetch("http://localhost:9090/api/processes");
      const psData = await psRes.json();
      setProcesses(psData);
    } catch (e) {
      console.error("Failed to fetch state:", e);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    const ws = new WebSocket("ws://localhost:9090/ws");
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.event === "log") {
          setLogs((prev) => [...prev.slice(-99), data.data]);
        }
      } catch (e) {
        setLogs((prev) => [...prev.slice(-99), String(event.data)]);
      }
    };
    return () => ws.close();
  }, []);

  const startProc = async (id: string, cmd: string, args: string[]) => {
    try {
      await fetch("http://localhost:9090/api/process/start", {
        method: "POST",
        body: JSON.stringify({ id, command: cmd, args }),
        headers: { "Content-Type": "application/json" }
      });
      fetchData();
    } catch (e) {
      console.error(e);
    }
  };

  const stopProc = async (id: string) => {
    try {
      await fetch("http://localhost:9090/api/process/stop", {
        method: "POST",
        body: JSON.stringify({ id }),
        headers: { "Content-Type": "application/json" }
      });
      fetchData();
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div style={{ padding: "20px", fontFamily: "Segoe UI, sans-serif", background: "#f0f2f5", minHeight: "100vh", display: "flex", flexDirection: "column", gap: "20px" }}>
      <header style={{ background: "#ffffff", padding: "20px", borderRadius: "10px", boxShadow: "0 2px 4px rgba(0,0,0,0.1)", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <h1 style={{ margin: 0, fontSize: "24px" }}>🧠 Brain Desktop Control Plane</h1>
        {status ? (
          <div style={{ textAlign: "right" }}>
            <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
              <div style={{ width: "12px", height: "12px", background: "#10b981", borderRadius: "50%" }}></div>
              <b>{status.status}</b>
            </div>
            <small style={{ color: "#6b7280" }}>{status.time}</small>
          </div>
        ) : (
          <div style={{ display: "flex", alignItems: "center", gap: "8px", color: "#ef4444" }}>
            <div style={{ width: "12px", height: "12px", background: "#ef4444", borderRadius: "50%" }}></div>
            <b>Disconnected</b>
          </div>
        )}
      </header>

      <div style={{ display: "flex", gap: "20px", flex: 1 }}>
        <section style={{ flex: 1, background: "#ffffff", padding: "20px", borderRadius: "10px", boxShadow: "0 2px 4px rgba(0,0,0,0.1)" }}>
          <h2 style={{ marginTop: 0 }}>⚙️ Managed Processes</h2>
          <div style={{ marginBottom: "20px", display: "flex", gap: "10px" }}>
             <button onClick={() => startProc("web_ping", "ping", ["-c", "5", "google.com"])} style={{ padding: "8px 16px", background: "#3b82f6", color: "white", border: "none", borderRadius: "6px", cursor: "pointer" }}>Start Diagnostics (Ping)</button>
          </div>
          <table style={{ width: "100%", borderCollapse: "collapse", textAlign: "left" }}>
            <thead>
              <tr style={{ borderBottom: "2px solid #e5e7eb" }}>
                <th style={{ padding: "10px" }}>ID</th>
                <th style={{ padding: "10px" }}>State</th>
                <th style={{ padding: "10px" }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {Object.entries(processes).map(([id, state]) => (
                <tr key={id} style={{ borderBottom: "1px solid #e5e7eb" }}>
                  <td style={{ padding: "10px", fontWeight: "500" }}>{id}</td>
                  <td style={{ padding: "10px", color: state === 'Running' ? '#10b981' : state === 'Failed' ? '#ef4444' : '#6b7280' }}>
                    {state}
                  </td>
                  <td style={{ padding: "10px" }}>
                    {state === 'Running' ? (
                      <button onClick={() => stopProc(id)} style={{ padding: "6px 12px", background: "#ef4444", color: "white", border: "none", borderRadius: "4px", cursor: "pointer" }}>Stop</button>
                    ) : (
                      <button disabled style={{ padding: "6px 12px", background: "#e5e7eb", color: "#9ca3af", border: "none", borderRadius: "4px" }}>Stopped</button>
                    )}
                  </td>
                </tr>
              ))}
              {Object.keys(processes).length === 0 && (
                <tr><td colSpan={3} style={{ padding: "20px", textAlign: "center", color: "#6b7280" }}>No processes running.</td></tr>
              )}
            </tbody>
          </table>
        </section>

        <section style={{ flex: 1, background: "#1f2937", color: "#f3f4f6", padding: "20px", borderRadius: "10px", boxShadow: "0 2px 4px rgba(0,0,0,0.1)", display: "flex", flexDirection: "column" }}>
          <h2 style={{ color: "#e5e7eb", margin: "0 0 10px 0" }}>⚡ Global Daemon Logs (Live)</h2>
          <div style={{ flex: 1, background: "#111827", padding: "15px", borderRadius: "6px", fontFamily: "monospace", overflowY: "auto", maxHeight: "500px", wordBreak: "break-all" }}>
            {logs.length === 0 ? <span style={{ color: "#4b5563" }}>Waiting for logs...</span> : logs.map((l, i) => (
              <div key={i} style={{ marginBottom: "4px", paddingBottom: "4px", borderBottom: "1px solid #1f2937" }}>{l}</div>
            ))}
            <div ref={logsEndRef} />
          </div>
        </section>
      </div>
    </div>
  );
}
