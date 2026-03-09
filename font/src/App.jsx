import React, { useMemo, useState } from "react";

function App() {
  const [apiBase, setApiBase] = useState("http://localhost:8080/api/v1");
  const [mode, setMode] = useState("login");
  const [token, setToken] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [deviceId, setDeviceId] = useState("dev_001");
  const [description, setDescription] = useState("audio upload");
  const [selectedFile, setSelectedFile] = useState(null);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  const isAuthed = useMemo(() => token.trim().length > 0, [token]);

  async function handleLogin(e) {
    e.preventDefault();
    setBusy(true);
    setMessage("");
    try {
      const res = await fetch(`${apiBase}/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password })
      });
      const json = await res.json();
      if (!res.ok || json.code !== 200) {
        throw new Error(json.message || "login failed");
      }
      setToken(json.data?.token || "");
      setMessage("登录成功");
    } catch (err) {
      setMessage(`登录失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleRegister(e) {
    e.preventDefault();
    setBusy(true);
    setMessage("");
    try {
      const res = await fetch(`${apiBase}/auth/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password, email })
      });
      const json = await res.json();
      if (!res.ok || json.code !== 200) {
        throw new Error(json.message || "register failed");
      }
      setMessage("注册成功，请切换到登录");
      setMode("login");
    } catch (err) {
      setMessage(`注册失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleUpload(e) {
    e.preventDefault();
    if (!selectedFile) {
      setMessage("请先选择文件");
      return;
    }
    setBusy(true);
    setMessage("");
    try {
      const fileName = selectedFile.name;
      const ext = fileName.includes(".") ? fileName.split(".").pop().toLowerCase() : "";
      const contentType = selectedFile.type || "application/octet-stream";

      const createRes = await fetch(`${apiBase}/jobs`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({
          device_id: deviceId,
          file_name: fileName,
          file_size: selectedFile.size,
          file_type: ext,
          content_type: contentType,
          description
        })
      });
      const createJson = await createRes.json();
      if (!createRes.ok || createJson.code !== 200) {
        throw new Error(createJson.message || "create upload job failed");
      }

      const job = createJson.data;
      const putRes = await fetch(job.upload_url, {
        method: "PUT",
        headers: {
          "Content-Type": contentType
        },
        body: selectedFile
      });
      if (!putRes.ok) {
        throw new Error(`minio upload failed: ${putRes.status}`);
      }

      const etag = (putRes.headers.get("etag") || "").replaceAll('"', "");
      const completeRes = await fetch(`${apiBase}/jobs/${job.job_id}/complete`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${token}`
        },
        body: JSON.stringify({
          job_id: job.job_id,
          bucket: job.bucket,
          key: job.key,
          etag,
          size: selectedFile.size,
          completed_at: new Date().toISOString()
        })
      });
      const completeJson = await completeRes.json();
      if (!completeRes.ok || completeJson.code !== 200) {
        throw new Error(completeJson.message || "complete callback failed");
      }

      setMessage(`上传成功，任务ID: ${job.job_id}`);
      setSelectedFile(null);
    } catch (err) {
      setMessage(`上传失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="page">
      <div className="card">
        <h1>RPW Detection</h1>
        <p className="sub">登录 / 注册 / 音频上传</p>

        <label className="label">API Base</label>
        <input value={apiBase} onChange={(e) => setApiBase(e.target.value)} />

        {!isAuthed && (
          <>
            <div className="tabs">
              <button className={mode === "login" ? "active" : ""} onClick={() => setMode("login")} type="button">
                登录
              </button>
              <button className={mode === "register" ? "active" : ""} onClick={() => setMode("register")} type="button">
                注册
              </button>
            </div>

            {mode === "login" ? (
              <form onSubmit={handleLogin} className="form">
                <input placeholder="用户名" value={username} onChange={(e) => setUsername(e.target.value)} />
                <input placeholder="密码" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
                <button disabled={busy} type="submit">
                  {busy ? "处理中..." : "登录"}
                </button>
              </form>
            ) : (
              <form onSubmit={handleRegister} className="form">
                <input placeholder="用户名" value={username} onChange={(e) => setUsername(e.target.value)} />
                <input placeholder="邮箱" value={email} onChange={(e) => setEmail(e.target.value)} />
                <input placeholder="密码" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
                <button disabled={busy} type="submit">
                  {busy ? "处理中..." : "注册"}
                </button>
              </form>
            )}
          </>
        )}

        {isAuthed && (
          <form onSubmit={handleUpload} className="form">
            <div className="token-line">Token 已保存（前 24 字符）: {token.slice(0, 24)}...</div>
            <input placeholder="设备ID" value={deviceId} onChange={(e) => setDeviceId(e.target.value)} />
            <input placeholder="描述" value={description} onChange={(e) => setDescription(e.target.value)} />
            <input
              type="file"
              accept=".wav,.mp3,.flac,.m4a,.aac,audio/*"
              onChange={(e) => setSelectedFile(e.target.files?.[0] || null)}
            />
            <button disabled={busy} type="submit">
              {busy ? "上传中..." : "上传文件"}
            </button>
            <button
              type="button"
              onClick={() => {
                setToken("");
                setMessage("已退出登录");
              }}
            >
              退出登录
            </button>
          </form>
        )}

        {message && <div className="msg">{message}</div>}
      </div>
    </div>
  );
}

export default App;
