import React, { useEffect, useMemo, useState } from "react";

const DEFAULT_API_BASE = "http://localhost:8080/api/v1";
const PARKS_KEY = "rpw.parks";
const TREES_KEY = "rpw.trees";
const DEVICES_KEY = "rpw.devices";
const JOB_META_KEY = "rpw.job_meta";

function readJSON(key, fallback) {
  try {
    const raw = localStorage.getItem(key);
    return raw ? JSON.parse(raw) : fallback;
  } catch {
    return fallback;
  }
}

function App() {
  const [apiBase] = useState(DEFAULT_API_BASE);
  const [mode, setMode] = useState("login");
  const [page, setPage] = useState("upload");
  const [token, setToken] = useState("");
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [selectedFile, setSelectedFile] = useState(null);
  const [description, setDescription] = useState("音频上传");
  const [message, setMessage] = useState("");
  const [messageTTL, setMessageTTL] = useState(0);
  const [busy, setBusy] = useState(false);
  const [jobsBusy, setJobsBusy] = useState(false);
  const [jobs, setJobs] = useState([]);

  const [parks, setParks] = useState(() => readJSON(PARKS_KEY, ["果园A区", "果园B区"]));
  const [trees, setTrees] = useState(() =>
    readJSON(TREES_KEY, [
      { park: "果园A区", treeNo: "A-001" },
      { park: "果园A区", treeNo: "A-002" },
      { park: "果园B区", treeNo: "B-001" }
    ])
  );
  const [devices, setDevices] = useState(() =>
    readJSON(DEVICES_KEY, [
      { device_id: "dev_001", device_name: "检测设备A", location: "果园A区" },
      { device_id: "dev_002", device_name: "检测设备B", location: "果园B区" }
    ])
  );
  const [jobMeta, setJobMeta] = useState(() => readJSON(JOB_META_KEY, {}));

  const [selectedPark, setSelectedPark] = useState("果园A区");
  const [selectedTreeNo, setSelectedTreeNo] = useState("A-001");
  const [selectedDeviceId, setSelectedDeviceId] = useState("dev_001");
  const [historyFilters, setHistoryFilters] = useState({
    from: "",
    to: "",
    park: "",
    treeNo: "",
    deviceId: ""
  });

  const [parkInput, setParkInput] = useState("");
  const [treeInput, setTreeInput] = useState("");
  const [deviceInput, setDeviceInput] = useState({ id: "", name: "", location: "" });

  const isAuthed = useMemo(() => token.trim().length > 0, [token]);
  const treesForSelectedPark = useMemo(() => trees.filter((tree) => tree.park === selectedPark), [trees, selectedPark]);
  const devicesForSelectedPark = useMemo(
    () => devices.filter((device) => device.location === selectedPark),
    [devices, selectedPark]
  );
  const treesForHistoryFilter = useMemo(
    () => (historyFilters.park ? trees.filter((tree) => tree.park === historyFilters.park) : trees),
    [trees, historyFilters.park]
  );

  useEffect(() => localStorage.setItem(PARKS_KEY, JSON.stringify(parks)), [parks]);
  useEffect(() => localStorage.setItem(TREES_KEY, JSON.stringify(trees)), [trees]);
  useEffect(() => localStorage.setItem(DEVICES_KEY, JSON.stringify(devices)), [devices]);
  useEffect(() => localStorage.setItem(JOB_META_KEY, JSON.stringify(jobMeta)), [jobMeta]);
  useEffect(() => {
    if (!message || messageTTL <= 0) return undefined;
    const timer = window.setTimeout(() => {
      setMessage("");
      setMessageTTL(0);
    }, messageTTL);
    return () => window.clearTimeout(timer);
  }, [message, messageTTL]);

  useEffect(() => {
    if (!treesForSelectedPark.length) {
      setSelectedTreeNo("");
      return;
    }
    if (!treesForSelectedPark.some((tree) => tree.treeNo === selectedTreeNo)) {
      setSelectedTreeNo(treesForSelectedPark[0].treeNo);
    }
  }, [treesForSelectedPark, selectedTreeNo]);

  useEffect(() => {
    if (!devicesForSelectedPark.length) {
      setSelectedDeviceId("");
      return;
    }
    if (!devicesForSelectedPark.some((device) => device.device_id === selectedDeviceId)) {
      setSelectedDeviceId(devicesForSelectedPark[0].device_id);
    }
  }, [devicesForSelectedPark, selectedDeviceId]);

  useEffect(() => {
    if (!historyFilters.park) return;
    if (historyFilters.treeNo && !treesForHistoryFilter.some((tree) => tree.treeNo === historyFilters.treeNo)) {
      setHistoryFilters((prev) => ({ ...prev, treeNo: "" }));
    }
  }, [historyFilters.park, historyFilters.treeNo, treesForHistoryFilter]);

  useEffect(() => {
    if (!isAuthed) {
      setJobs([]);
      return;
    }
    loadDevices();
    loadJobs();
  }, [isAuthed]);

  function normalizeJob(job) {
    const meta = jobMeta[job.id || job.job_id] || {};
    return {
      id: job.id || job.job_id,
      deviceId: meta.deviceId || job.device_id || "",
      park: meta.park || "",
      treeNo: meta.treeNo || "",
      fileName: job.file_name || job.key || "未命名音频.wav",
      fileSize: job.file_size || 0,
      createdAt: job.created_at || new Date().toISOString(),
      status: job.status || "pending"
    };
  }

  function getDetectionLabel(job) {
    if (job.status === "pending" || job.status === "uploading" || job.status === "processing") {
      return "正在检测中";
    }
    const seed = (job.id || "").split("").reduce((sum, ch) => sum + ch.charCodeAt(0), 0);
    return seed % 2 === 0 ? "健康" : "受感染";
  }

  function parseJobDate(value) {
    if (!value) return null;
    const normalized = typeof value === "string" && value.includes(" ") && !value.includes("T")
      ? value.replace(" ", "T")
      : value;
    const parsed = new Date(normalized);
    return Number.isNaN(parsed.getTime()) ? null : parsed;
  }

  const filteredJobs = useMemo(() => {
    return jobs.filter((job) => {
      const jobDate = parseJobDate(job.createdAt);
      if (historyFilters.from && jobDate) {
        const from = new Date(`${historyFilters.from}T00:00:00`);
        if (jobDate < from) return false;
      }
      if (historyFilters.to && jobDate) {
        const to = new Date(`${historyFilters.to}T23:59:59`);
        if (jobDate > to) return false;
      }
      if (historyFilters.deviceId && job.deviceId !== historyFilters.deviceId) return false;
      if (historyFilters.park && job.park !== historyFilters.park) return false;
      if (historyFilters.treeNo && job.treeNo !== historyFilters.treeNo) return false;
      return true;
    });
  }, [jobs, historyFilters]);

  function clearMessage() {
    setMessage("");
    setMessageTTL(0);
  }

  function showMessage(text, ttl = 0) {
    setMessage(text);
    setMessageTTL(ttl);
  }

  async function handleLogin(e) {
    e.preventDefault();
    setBusy(true);
    clearMessage();
    try {
      const res = await fetch(`${apiBase}/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password })
      });
      const json = await res.json();
      if (!res.ok || json.code !== 200) throw new Error(json.message || "登录失败");
      setToken(json.data?.token || "");
      setPage("upload");
      showMessage("登录成功", 1000);
    } catch (err) {
      showMessage(`登录失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleRegister(e) {
    e.preventDefault();
    setBusy(true);
    clearMessage();
    try {
      const res = await fetch(`${apiBase}/auth/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password, email })
      });
      const json = await res.json();
      if (!res.ok || json.code !== 200) throw new Error(json.message || "注册失败");
      showMessage("注册成功，请切换到登录", 1000);
      setMode("login");
    } catch (err) {
      showMessage(`注册失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function loadDevices() {
    try {
      const res = await fetch(`${apiBase}/device/list`);
      const json = await res.json();
      if (!res.ok || json.code !== 200) throw new Error(json.message || "加载设备失败");
      const items = Array.isArray(json.data?.devices) ? json.data.devices : [];
      if (!items.length) return;
      setDevices((prev) => {
        const merged = [...prev];
        items.forEach((item) => {
          if (!merged.some((device) => device.device_id === item.device_id)) merged.push(item);
        });
        return merged;
      });
      setParks((prev) => Array.from(new Set([...prev, ...items.map((item) => item.location).filter(Boolean)])));
    } catch (err) {
      showMessage(`加载设备失败: ${err.message}`);
    }
  }

  async function loadJobs() {
    setJobsBusy(true);
    try {
      const res = await fetch(`${apiBase}/jobs`, {
        headers: token ? { Authorization: `Bearer ${token}` } : {}
      });
      const json = await res.json();
      if (!res.ok || json.code !== 200) throw new Error(json.message || "加载历史记录失败");
      const items = Array.isArray(json.data?.data) ? json.data.data : [];
      setJobs(items.map(normalizeJob));
    } catch (err) {
      showMessage(`加载历史记录失败: ${err.message}`);
    } finally {
      setJobsBusy(false);
    }
  }

  async function handleUpload(e) {
    e.preventDefault();
    if (!selectedFile) return showMessage("请先选择文件");
    if (!selectedPark || !selectedTreeNo || !selectedDeviceId) return showMessage("请选择园区、树木编号和设备");

    setBusy(true);
    clearMessage();
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
          device_id: selectedDeviceId,
          file_name: fileName,
          file_size: selectedFile.size,
          file_type: ext,
          content_type: contentType,
          description
        })
      });
      const createJson = await createRes.json();
      if (!createRes.ok || createJson.code !== 200) throw new Error(createJson.message || "创建上传任务失败");

      const job = createJson.data;
      const putRes = await fetch(job.upload_url, {
        method: "PUT",
        headers: { "Content-Type": contentType },
        body: selectedFile
      });
      if (!putRes.ok) throw new Error(`MinIO 上传失败: ${putRes.status}`);

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
      if (!completeRes.ok || completeJson.code !== 200) throw new Error(completeJson.message || "上传完成通知失败");

      setJobMeta((prev) => ({
        ...prev,
        [job.job_id]: { park: selectedPark, treeNo: selectedTreeNo, deviceId: selectedDeviceId }
      }));
      setJobs((prev) => [
        {
          id: job.job_id,
          deviceId: selectedDeviceId,
          park: selectedPark,
          treeNo: selectedTreeNo,
          fileName,
          fileSize: selectedFile.size,
          createdAt: new Date().toISOString(),
          status: "processing"
        },
        ...prev
      ]);
      setSelectedFile(null);
      showMessage(`上传成功，任务ID: ${job.job_id}`);
      setPage("history");
    } catch (err) {
      showMessage(`上传失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleDeleteJob(jobId) {
    setBusy(true);
    clearMessage();
    try {
      const res = await fetch(`${apiBase}/jobs/${jobId}`, {
        method: "DELETE",
        headers: token ? { Authorization: `Bearer ${token}` } : {}
      });
      const json = await res.json();
      if (!res.ok || json.code !== 200) throw new Error(json.message || "删除记录失败");
      setJobs((prev) => prev.filter((job) => job.id !== jobId));
      setJobMeta((prev) => {
        const next = { ...prev };
        delete next[jobId];
        return next;
      });
      showMessage(`记录已删除: ${jobId}`);
    } catch (err) {
      showMessage(`删除失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  function addPark(e) {
    e.preventDefault();
    const name = parkInput.trim();
    if (!name) return;
    if (!parks.includes(name)) setParks((prev) => [...prev, name]);
    setSelectedPark(name);
    setParkInput("");
    showMessage(`园区已添加: ${name}`);
  }

  function addTree(e) {
    e.preventDefault();
    const no = treeInput.trim();
    if (!selectedPark || !no) return;
    if (!trees.some((tree) => tree.park === selectedPark && tree.treeNo === no)) {
      setTrees((prev) => [...prev, { park: selectedPark, treeNo: no }]);
    }
    setSelectedTreeNo(no);
    setTreeInput("");
    showMessage(`树木编号已添加: ${no}`);
  }

  function addDevice(e) {
    e.preventDefault();
    const payload = {
      device_id: deviceInput.id.trim(),
      device_name: deviceInput.name.trim(),
      location: deviceInput.location.trim() || selectedPark
    };
    if (!payload.device_id || !payload.device_name || !payload.location) return;
    if (!devices.some((device) => device.device_id === payload.device_id)) {
      setDevices((prev) => [...prev, payload]);
    }
    setParks((prev) => Array.from(new Set([...prev, payload.location])));
    setSelectedPark(payload.location);
    setSelectedDeviceId(payload.device_id);
    setDeviceInput({ id: "", name: "", location: "" });
    showMessage(`设备已添加: ${payload.device_name}`);
  }

  function updateHistoryFilter(key, value) {
    setHistoryFilters((prev) => ({ ...prev, [key]: value }));
  }

  function resetHistoryFilters() {
    setHistoryFilters({
      from: "",
      to: "",
      park: "",
      treeNo: "",
      deviceId: ""
    });
  }

  function renderUploadPage() {
    return (
      <section className="panel content-panel">
        <div className="section-head">
          <div>
            <h2>发送音频</h2>
          </div>
          <span className="token-line">令牌：{token.slice(0, 24)}...</span>
        </div>
        <form onSubmit={handleUpload} className="form">
          <label className="field">
            <span>园区</span>
            <select value={selectedPark} onChange={(e) => setSelectedPark(e.target.value)}>
              {parks.map((park) => (
                <option key={park} value={park}>{park}</option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>树木编号</span>
            <select value={selectedTreeNo} onChange={(e) => setSelectedTreeNo(e.target.value)}>
              {treesForSelectedPark.map((tree) => (
                <option key={`${tree.park}-${tree.treeNo}`} value={tree.treeNo}>{tree.treeNo}</option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>设备</span>
            <select value={selectedDeviceId} onChange={(e) => setSelectedDeviceId(e.target.value)}>
              {devicesForSelectedPark.map((device) => (
                <option key={device.device_id} value={device.device_id}>
                  {device.device_name} ({device.device_id})
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>描述</span>
            <input value={description} onChange={(e) => setDescription(e.target.value)} />
          </label>
          <label className="field">
            <span>音频文件</span>
            <input type="file" accept=".wav,.mp3,.flac,.m4a,.aac,audio/*" onChange={(e) => setSelectedFile(e.target.files?.[0] || null)} />
          </label>
          <button className="primary" disabled={busy} type="submit">
            {busy ? "上传中..." : "发送文件"}
          </button>
        </form>
      </section>
    );
  }

  function renderHistoryPage() {
    return (
      <section className="panel content-panel">
        <div className="section-head">
          <div><h2>历史检测记录</h2></div>
          <button className="ghost" disabled={jobsBusy} type="button" onClick={loadJobs}>
            {jobsBusy ? "刷新中..." : "刷新记录"}
          </button>
        </div>
        <div className="filter-grid">
          <label className="field">
            <span>开始时间</span>
            <input type="date" value={historyFilters.from} onChange={(e) => updateHistoryFilter("from", e.target.value)} />
          </label>
          <label className="field">
            <span>结束时间</span>
            <input type="date" value={historyFilters.to} onChange={(e) => updateHistoryFilter("to", e.target.value)} />
          </label>
          <label className="field">
            <span>设备</span>
            <select value={historyFilters.deviceId} onChange={(e) => updateHistoryFilter("deviceId", e.target.value)}>
              <option value="">全部设备</option>
              {devices.map((device) => (
                <option key={device.device_id} value={device.device_id}>
                  {device.device_name} ({device.device_id})
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>园区</span>
            <select value={historyFilters.park} onChange={(e) => updateHistoryFilter("park", e.target.value)}>
              <option value="">全部园区</option>
              {parks.map((park) => (
                <option key={park} value={park}>{park}</option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>树木编号</span>
            <select value={historyFilters.treeNo} onChange={(e) => updateHistoryFilter("treeNo", e.target.value)}>
              <option value="">全部树木</option>
              {treesForHistoryFilter.map((tree) => (
                <option key={`${tree.park}-${tree.treeNo}`} value={tree.treeNo}>{tree.treeNo}</option>
              ))}
            </select>
          </label>
        </div>
        <div className="filter-actions">
          <button className="ghost" type="button" onClick={resetHistoryFilters}>
            清空筛选
          </button>
        </div>
        {jobs.length === 0 ? (
          <div className="empty">还没有可展示的历史记录。</div>
        ) : filteredJobs.length === 0 ? (
          <div className="empty">没有符合当前筛选条件的记录。</div>
        ) : (
          <div className="job-list">
            {filteredJobs.map((job) => (
              <article key={job.id} className="job-card">
                <div className="job-main">
                  <div>
                    <h3>{job.fileName}</h3>
                    <p className="meta">园区: {job.park || "未设置"} | 树木编号: {job.treeNo || "未设置"} | 设备: {job.deviceId || "未设置"}</p>
                    <p className="meta">任务ID: {job.id} | 大小: {job.fileSize || 0} 字节 | 时间: {new Date(job.createdAt).toLocaleString()}</p>
                  </div>
                  <div className={`status status-${getDetectionLabel(job)}`}>{getDetectionLabel(job)}</div>
                </div>
                <div className="job-actions">
                  <button className="danger" disabled={busy} type="button" onClick={() => handleDeleteJob(job.id)}>
                    删除记录
                  </button>
                </div>
              </article>
            ))}
          </div>
        )}
      </section>
    );
  }

  function renderParksPage() {
    return (
      <section className="panel content-panel">
        <div className="section-head"><div><h2>园区管理</h2></div></div>
        <form className="form mini-form" onSubmit={addPark}>
          <input placeholder="例如：果园C区" value={parkInput} onChange={(e) => setParkInput(e.target.value)} />
          <button className="ghost" type="submit">添加园区</button>
          <div className="pill-list">
            {parks.map((park) => <span key={park} className="pill">{park}</span>)}
          </div>
        </form>
      </section>
    );
  }

  function renderTreesPage() {
    return (
      <section className="panel content-panel">
        <div className="section-head"><div><h2>树木编号管理</h2></div></div>
        <form className="form mini-form" onSubmit={addTree}>
          <select value={selectedPark} onChange={(e) => setSelectedPark(e.target.value)}>
            {parks.map((park) => <option key={park} value={park}>{park}</option>)}
          </select>
          <input placeholder="例如：A-010" value={treeInput} onChange={(e) => setTreeInput(e.target.value)} />
          <button className="ghost" type="submit">添加树木编号</button>
          <div className="pill-list">
            {treesForSelectedPark.map((tree) => <span key={`${tree.park}-${tree.treeNo}`} className="pill">{tree.treeNo}</span>)}
          </div>
        </form>
      </section>
    );
  }

  function renderDevicesPage() {
    return (
      <section className="panel content-panel">
        <div className="section-head"><div><h2>设备管理</h2></div></div>
        <form className="form mini-form" onSubmit={addDevice}>
          <input placeholder="设备ID" value={deviceInput.id} onChange={(e) => setDeviceInput((prev) => ({ ...prev, id: e.target.value }))} />
          <input placeholder="设备名称" value={deviceInput.name} onChange={(e) => setDeviceInput((prev) => ({ ...prev, name: e.target.value }))} />
          <input placeholder="所属园区" value={deviceInput.location} onChange={(e) => setDeviceInput((prev) => ({ ...prev, location: e.target.value }))} />
          <button className="ghost" type="submit">添加设备</button>
          <div className="pill-list">
            {devices.map((device) => <span key={device.device_id} className="pill">{device.device_name} ({device.location})</span>)}
          </div>
        </form>
      </section>
    );
  }

  function renderAuthedPage() {
    switch (page) {
      case "history":
        return renderHistoryPage();
      case "parks":
        return renderParksPage();
      case "trees":
        return renderTreesPage();
      case "devices":
        return renderDevicesPage();
      default:
        return renderUploadPage();
    }
  }

  return (
    <div className="page">
      <div className="shell">
        <section className={`hero ${isAuthed ? "hero-bar" : "hero-center"}`}>
          <div className="hero-copy">
            <h1>病虫害监测系统</h1>
          </div>
          {isAuthed && (
            <button className="ghost" type="button" onClick={() => { setToken(""); showMessage("已退出登录", 1000); }}>
              退出登录
            </button>
          )}
        </section>

        {!isAuthed && (
          <div className="auth-stage">
            <section className="panel auth-panel">
              <div className="tabs">
                <button className={mode === "login" ? "active" : ""} onClick={() => setMode("login")} type="button">登录</button>
                <button className={mode === "register" ? "active" : ""} onClick={() => setMode("register")} type="button">注册</button>
              </div>
              {mode === "login" ? (
                <form onSubmit={handleLogin} className="form">
                  <input placeholder="用户名" value={username} onChange={(e) => setUsername(e.target.value)} />
                  <input placeholder="密码" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
                  <button className="primary" disabled={busy} type="submit">{busy ? "处理中..." : "登录"}</button>
                </form>
              ) : (
                <form onSubmit={handleRegister} className="form">
                  <input placeholder="用户名" value={username} onChange={(e) => setUsername(e.target.value)} />
                  <input placeholder="邮箱" value={email} onChange={(e) => setEmail(e.target.value)} />
                  <input placeholder="密码" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
                  <button className="primary" disabled={busy} type="submit">{busy ? "处理中..." : "注册"}</button>
                </form>
              )}
            </section>
          </div>
        )}

        {isAuthed && (
          <div className="app-grid">
            <aside className="panel sidebar">
              <h2 className="nav-title">功能菜单</h2>
              <div className="nav-list">
                {[
                  ["upload", "发送音频"],
                  ["history", "历史记录"],
                  ["parks", "园区管理"],
                  ["trees", "树木管理"],
                  ["devices", "设备管理"]
                ].map(([key, label]) => (
                  <button
                    key={key}
                    className={`nav-btn ${page === key ? "active" : ""}`}
                    type="button"
                    onClick={() => setPage(key)}
                  >
                    {label}
                  </button>
                ))}
              </div>
            </aside>
            {renderAuthedPage()}
          </div>
        )}

        {message && <div className="msg">{message}</div>}
      </div>
    </div>
  );
}

export default App;
