import React, { useEffect, useMemo, useState } from "react";

const DEFAULT_API_BASE = "http://localhost:8080/api/v1";
const TOKEN_STORAGE_KEY = "rpw_token";
const USER_STORAGE_KEY = "rpw_user";
const PAGE_STORAGE_KEY = "rpw_page";

function readStoredToken() {
  if (typeof window === "undefined") return "";
  return window.localStorage.getItem(TOKEN_STORAGE_KEY) || "";
}

function readStoredUser() {
  if (typeof window === "undefined") return null;
  const raw = window.localStorage.getItem(USER_STORAGE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch (_err) {
    window.localStorage.removeItem(USER_STORAGE_KEY);
    return null;
  }
}

function readStoredPage() {
  if (typeof window === "undefined") return "upload";
  const page = window.localStorage.getItem(PAGE_STORAGE_KEY) || "upload";
  return ["upload", "history", "parks", "trees", "devices"].includes(page) ? page : "upload";
}

function parseDate(value) {
  if (!value) return null;
  const normalized = typeof value === "string" && value.includes(" ") && !value.includes("T")
    ? value.replace(" ", "T")
    : value;
  const parsed = new Date(normalized);
  return Number.isNaN(parsed.getTime()) ? null : parsed;
}

function App() {
  const [apiBase] = useState(DEFAULT_API_BASE);
  const [mode, setMode] = useState("login");
  const [page, setPage] = useState(readStoredPage);
  const [token, setToken] = useState(readStoredToken);
  const [user, setUser] = useState(readStoredUser);
  const [username, setUsername] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [selectedFile, setSelectedFile] = useState(null);
  const [description, setDescription] = useState("音频上传");
  const [message, setMessage] = useState("");
  const [messageTTL, setMessageTTL] = useState(0);
  const [busy, setBusy] = useState(false);
  const [jobsBusy, setJobsBusy] = useState(false);
  const [referenceBusy, setReferenceBusy] = useState(false);

  const [parks, setParks] = useState([]);
  const [trees, setTrees] = useState([]);
  const [devices, setDevices] = useState([]);
  const [jobs, setJobs] = useState([]);

  const [selectedParkId, setSelectedParkId] = useState("");
  const [selectedTreeId, setSelectedTreeId] = useState("");
  const [selectedDeviceCode, setSelectedDeviceCode] = useState("");

  const [historyFilters, setHistoryFilters] = useState({
    from: "",
    to: "",
    parkId: "",
    treeId: "",
    deviceCode: ""
  });

  const [parkForm, setParkForm] = useState({ name: "", code: "", location: "" });
  const [treeForm, setTreeForm] = useState({ parkId: "", treeCode: "", species: "" });
  const [deviceForm, setDeviceForm] = useState({ parkId: "", treeId: "", deviceCode: "", name: "" });

  const isAuthed = token.trim().length > 0;
  const isAdmin = user?.is_admin === 1;

  const treesForSelectedPark = useMemo(
    () => trees.filter((tree) => tree.parkId === selectedParkId),
    [trees, selectedParkId]
  );

  const devicesForSelectedTree = useMemo(() => {
    return devices.filter((device) => {
      if (selectedParkId && device.parkId !== selectedParkId) return false;
      if (selectedTreeId && device.treeId !== selectedTreeId) return false;
      return true;
    });
  }, [devices, selectedParkId, selectedTreeId]);

  const treesForHistoryFilter = useMemo(() => {
    if (!historyFilters.parkId) return trees;
    return trees.filter((tree) => tree.parkId === historyFilters.parkId);
  }, [trees, historyFilters.parkId]);

  const devicesForHistoryFilter = useMemo(() => {
    return devices.filter((device) => {
      if (historyFilters.parkId && device.parkId !== historyFilters.parkId) return false;
      if (historyFilters.treeId && device.treeId !== historyFilters.treeId) return false;
      return true;
    });
  }, [devices, historyFilters.parkId, historyFilters.treeId]);

  const treesForDeviceForm = useMemo(
    () => trees.filter((tree) => tree.parkId === deviceForm.parkId),
    [trees, deviceForm.parkId]
  );

  const filteredJobs = useMemo(() => {
    return jobs.filter((job) => {
      const createdAt = parseDate(job.createdAt);
      if (historyFilters.from && createdAt) {
        const from = new Date(`${historyFilters.from}T00:00:00`);
        if (createdAt < from) return false;
      }
      if (historyFilters.to && createdAt) {
        const to = new Date(`${historyFilters.to}T23:59:59`);
        if (createdAt > to) return false;
      }
      if (historyFilters.parkId && job.parkId !== historyFilters.parkId) return false;
      if (historyFilters.treeId && job.treeId !== historyFilters.treeId) return false;
      if (historyFilters.deviceCode && job.deviceCode !== historyFilters.deviceCode) return false;
      return true;
    });
  }, [jobs, historyFilters]);

  useEffect(() => {
    if (!message || messageTTL <= 0) return undefined;
    const timer = window.setTimeout(() => {
      setMessage("");
      setMessageTTL(0);
    }, messageTTL);
    return () => window.clearTimeout(timer);
  }, [message, messageTTL]);

  useEffect(() => {
    if (typeof window === "undefined") return;
    if (token) {
      window.localStorage.setItem(TOKEN_STORAGE_KEY, token);
    } else {
      window.localStorage.removeItem(TOKEN_STORAGE_KEY);
    }
  }, [token]);

  useEffect(() => {
    if (typeof window === "undefined") return;
    if (user) {
      window.localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(user));
    } else {
      window.localStorage.removeItem(USER_STORAGE_KEY);
    }
  }, [user]);

  useEffect(() => {
    if (typeof window === "undefined") return;
    if (page) {
      window.localStorage.setItem(PAGE_STORAGE_KEY, page);
    } else {
      window.localStorage.removeItem(PAGE_STORAGE_KEY);
    }
  }, [page]);

  useEffect(() => {
    if (!isAuthed) {
      setParks([]);
      setTrees([]);
      setDevices([]);
      setJobs([]);
      setUser(null);
      return;
    }
    void bootstrap();
  }, [isAuthed]);

  useEffect(() => {
    if (!isAdmin && ["parks", "trees", "devices"].includes(page)) {
      setPage("upload");
    }
  }, [isAdmin, page]);

  useEffect(() => {
    if (!parks.length) {
      setSelectedParkId("");
      setTreeForm((prev) => ({ ...prev, parkId: "" }));
      setDeviceForm((prev) => ({ ...prev, parkId: "", treeId: "" }));
      return;
    }
    if (!parks.some((park) => park.id === selectedParkId)) {
      setSelectedParkId(parks[0].id);
    }
    setTreeForm((prev) => (prev.parkId && parks.some((park) => park.id === prev.parkId) ? prev : { ...prev, parkId: parks[0].id }));
    setDeviceForm((prev) => (prev.parkId && parks.some((park) => park.id === prev.parkId) ? prev : { ...prev, parkId: parks[0].id, treeId: "" }));
  }, [parks, selectedParkId]);

  useEffect(() => {
    if (!treesForSelectedPark.length) {
      setSelectedTreeId("");
      return;
    }
    if (!treesForSelectedPark.some((tree) => tree.id === selectedTreeId)) {
      setSelectedTreeId(treesForSelectedPark[0].id);
    }
  }, [treesForSelectedPark, selectedTreeId]);

  useEffect(() => {
    if (!treesForDeviceForm.length) {
      setDeviceForm((prev) => ({ ...prev, treeId: "" }));
      return;
    }
    if (!treesForDeviceForm.some((tree) => tree.id === deviceForm.treeId)) {
      setDeviceForm((prev) => ({ ...prev, treeId: treesForDeviceForm[0].id }));
    }
  }, [treesForDeviceForm, deviceForm.treeId]);

  useEffect(() => {
    if (!devicesForSelectedTree.length) {
      setSelectedDeviceCode("");
      return;
    }
    if (!devicesForSelectedTree.some((device) => device.deviceCode === selectedDeviceCode)) {
      setSelectedDeviceCode(devicesForSelectedTree[0].deviceCode);
    }
  }, [devicesForSelectedTree, selectedDeviceCode]);

  useEffect(() => {
    if (!historyFilters.parkId) return;
    if (historyFilters.treeId && !treesForHistoryFilter.some((tree) => tree.id === historyFilters.treeId)) {
      setHistoryFilters((prev) => ({ ...prev, treeId: "" }));
    }
    if (historyFilters.deviceCode && !devicesForHistoryFilter.some((device) => device.deviceCode === historyFilters.deviceCode)) {
      setHistoryFilters((prev) => ({ ...prev, deviceCode: "" }));
    }
  }, [historyFilters.parkId, historyFilters.treeId, historyFilters.deviceCode, treesForHistoryFilter, devicesForHistoryFilter]);

  function showMessage(text, ttl = 0) {
    setMessage(text);
    setMessageTTL(ttl);
  }

  function clearMessage() {
    setMessage("");
    setMessageTTL(0);
  }

  function authHeaders(includeJSON = false) {
    const headers = {};
    if (includeJSON) headers["Content-Type"] = "application/json";
    if (token) headers.Authorization = `Bearer ${token}`;
    return headers;
  }

  async function requestJSON(path, options = {}) {
    const res = await fetch(`${apiBase}${path}`, options);
    const json = await res.json().catch(() => ({}));
    if (!res.ok || json.code !== 200) {
      throw new Error(json.message || `请求失败: ${res.status}`);
    }
    return json.data;
  }

  function normalizePark(park) {
    return {
      id: String(park.id),
      name: park.name || "",
      code: park.code || "",
      location: park.location || "",
      status: Number(park.status ?? 1)
    };
  }

  function normalizeTree(tree) {
    return {
      id: String(tree.id),
      parkId: String(tree.park_id),
      parkName: tree.park?.name || "",
      treeCode: tree.tree_code || "",
      species: tree.species || "",
      status: Number(tree.status ?? 1)
    };
  }

  function normalizeDevice(device) {
    return {
      id: String(device.id),
      deviceCode: device.device_code || device.device_id || "",
      deviceName: device.device_name || device.name || "",
      parkId: String(device.park_id),
      parkName: device.park_name || device.location || "",
      treeId: device.tree_id == null ? "" : String(device.tree_id),
      treeCode: device.tree_code || "",
      status: Number(device.status ?? 1)
    };
  }

  function normalizeJob(job) {
    return {
      id: job.id || job.job_id,
      jobId: job.job_id || job.id,
      parkId: String(job.park_id ?? ""),
      parkName: job.park_name || "",
      treeId: String(job.tree_id ?? ""),
      treeCode: job.tree_code || "",
      deviceCode: job.device_code || job.device_id || "",
      deviceName: job.device_name || "",
      fileName: job.file_name || "未命名音频.wav",
      fileSize: Number(job.file_size || 0),
      status: job.status || "pending",
      resultLabel: job.result_label || "",
      resultScore: job.result_score == null ? null : Number(job.result_score),
      errorMessage: job.error_message || "",
      createdAt: job.created_at || new Date().toISOString()
    };
  }

  function getDetectionLabel(job) {
    if (job.status === "failed") return "检测失败";
    if (job.resultLabel === "clean") return "健康";
    if (job.resultLabel === "infested") return "受感染";
    if (job.status === "done") return "已完成";
    return "正在检测中";
  }

  async function bootstrap() {
    setReferenceBusy(true);
    try {
      await Promise.all([
        loadParks(true),
        loadTrees(true),
        loadDevices(true),
        loadJobs(true)
      ]);
    } catch (err) {
      showMessage(`加载业务数据失败: ${err.message}`);
    } finally {
      setReferenceBusy(false);
    }
  }

  async function loadParks(silent = false) {
    try {
      const data = await requestJSON("/parks", {
        headers: authHeaders()
      });
      const items = Array.isArray(data?.parks) ? data.parks.map(normalizePark) : [];
      setParks(items);
    } catch (err) {
      if (!silent) showMessage(`加载园区失败: ${err.message}`);
      throw err;
    }
  }

  async function loadTrees(silent = false) {
    try {
      const data = await requestJSON("/trees", {
        headers: authHeaders()
      });
      const items = Array.isArray(data?.trees) ? data.trees.map(normalizeTree) : [];
      setTrees(items);
    } catch (err) {
      if (!silent) showMessage(`加载树木失败: ${err.message}`);
      throw err;
    }
  }

  async function loadDevices(silent = false) {
    try {
      const data = await requestJSON("/devices", {
        headers: authHeaders()
      });
      const items = Array.isArray(data?.devices) ? data.devices.map(normalizeDevice) : [];
      setDevices(items);
    } catch (err) {
      if (!silent) showMessage(`加载设备失败: ${err.message}`);
      throw err;
    }
  }

  async function loadJobs(silent = false) {
    setJobsBusy(true);
    try {
      const data = await requestJSON("/jobs", {
        headers: authHeaders()
      });
      const items = Array.isArray(data?.data) ? data.data.map(normalizeJob) : [];
      setJobs(items);
    } catch (err) {
      if (!silent) showMessage(`加载历史记录失败: ${err.message}`);
      throw err;
    } finally {
      setJobsBusy(false);
    }
  }

  async function handleLogin(e) {
    e.preventDefault();
    setBusy(true);
    clearMessage();
    try {
      const data = await requestJSON("/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password })
      });
      setToken(data?.token || "");
      setUser(data?.user || null);
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
      await requestJSON("/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password, email })
      });
      showMessage("注册成功，请切换到登录", 1000);
      setMode("login");
    } catch (err) {
      showMessage(`注册失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleUpload(e) {
    e.preventDefault();
    if (!selectedFile) {
      showMessage("请先选择文件");
      return;
    }
    if (!selectedParkId || !selectedTreeId || !selectedDeviceCode) {
      showMessage("请选择园区、树木和设备");
      return;
    }

    setBusy(true);
    clearMessage();
    try {
      const fileName = selectedFile.name;
      const ext = fileName.includes(".") ? fileName.split(".").pop().toLowerCase() : "";
      const contentType = selectedFile.type || "application/octet-stream";
      const job = await requestJSON("/jobs", {
        method: "POST",
        headers: authHeaders(true),
        body: JSON.stringify({
          device_id: selectedDeviceCode,
          file_name: fileName,
          file_size: selectedFile.size,
          file_type: ext,
          content_type: contentType,
          description
        })
      });

      const putRes = await fetch(job.upload_url, {
        method: "PUT",
        headers: { "Content-Type": contentType },
        body: selectedFile
      });
      if (!putRes.ok) {
        throw new Error(`MinIO 上传失败: ${putRes.status}`);
      }

      const etag = (putRes.headers.get("etag") || "").replaceAll('"', "");
      await requestJSON(`/jobs/${job.job_id}/complete`, {
        method: "POST",
        headers: authHeaders(true),
        body: JSON.stringify({
          job_id: job.job_id,
          bucket: job.bucket,
          key: job.key,
          etag,
          size: selectedFile.size,
          completed_at: new Date().toISOString()
        })
      });

      setSelectedFile(null);
      await loadJobs(true);
      showMessage(`上传成功，任务ID: ${job.job_id}`, 1500);
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
      await requestJSON(`/jobs/${jobId}`, {
        method: "DELETE",
        headers: authHeaders()
      });
      await loadJobs(true);
      showMessage(`记录已删除: ${jobId}`, 1200);
    } catch (err) {
      showMessage(`删除失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleCreatePark(e) {
    e.preventDefault();
    if (!parkForm.name.trim() || !parkForm.code.trim()) {
      showMessage("请填写园区名称和编码");
      return;
    }

    setBusy(true);
    clearMessage();
    try {
      const created = await requestJSON("/parks", {
        method: "POST",
        headers: authHeaders(true),
        body: JSON.stringify({
          name: parkForm.name.trim(),
          code: parkForm.code.trim(),
          location: parkForm.location.trim()
        })
      });
      await loadParks(true);
      setSelectedParkId(String(created.id));
      setParkForm({ name: "", code: "", location: "" });
      showMessage(`园区已添加: ${created.name}`, 1200);
    } catch (err) {
      showMessage(`添加园区失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleCreateTree(e) {
    e.preventDefault();
    if (!treeForm.parkId || !treeForm.treeCode.trim()) {
      showMessage("请选择园区并填写树木编号");
      return;
    }

    setBusy(true);
    clearMessage();
    try {
      const created = await requestJSON("/trees", {
        method: "POST",
        headers: authHeaders(true),
        body: JSON.stringify({
          park_id: Number(treeForm.parkId),
          tree_code: treeForm.treeCode.trim(),
          species: treeForm.species.trim()
        })
      });
      await loadTrees(true);
      setSelectedParkId(String(created.park_id));
      setSelectedTreeId(String(created.id));
      setTreeForm((prev) => ({ ...prev, treeCode: "", species: "" }));
      showMessage(`树木已添加: ${created.tree_code}`, 1200);
    } catch (err) {
      showMessage(`添加树木失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleCreateDevice(e) {
    e.preventDefault();
    if (!deviceForm.parkId || !deviceForm.treeId || !deviceForm.deviceCode.trim() || !deviceForm.name.trim()) {
      showMessage("请填写完整的设备信息");
      return;
    }

    setBusy(true);
    clearMessage();
    try {
      const created = await requestJSON("/devices", {
        method: "POST",
        headers: authHeaders(true),
        body: JSON.stringify({
          device_code: deviceForm.deviceCode.trim(),
          name: deviceForm.name.trim(),
          park_id: Number(deviceForm.parkId),
          tree_id: Number(deviceForm.treeId)
        })
      });
      await loadDevices(true);
      setSelectedParkId(String(created.park_id));
      if (created.tree_id != null) setSelectedTreeId(String(created.tree_id));
      setSelectedDeviceCode(created.device_code);
      setDeviceForm((prev) => ({ ...prev, deviceCode: "", name: "" }));
      showMessage(`设备已添加: ${created.device_name}`, 1200);
    } catch (err) {
      showMessage(`添加设备失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleDeletePark(parkId) {
    if (!window.confirm("确定删除这个园区吗？")) return;
    setBusy(true);
    clearMessage();
    try {
      await requestJSON(`/parks/${parkId}`, {
        method: "DELETE",
        headers: authHeaders()
      });
      await Promise.all([loadParks(true), loadTrees(true), loadDevices(true)]);
      showMessage("园区已删除", 1200);
    } catch (err) {
      showMessage(`删除园区失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleDeleteTree(treeId) {
    if (!window.confirm("确定删除这棵树吗？")) return;
    setBusy(true);
    clearMessage();
    try {
      await requestJSON(`/trees/${treeId}`, {
        method: "DELETE",
        headers: authHeaders()
      });
      await Promise.all([loadTrees(true), loadDevices(true)]);
      showMessage("树木已删除", 1200);
    } catch (err) {
      showMessage(`删除树木失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  async function handleDeleteDevice(deviceId) {
    if (!window.confirm("确定删除这个设备吗？")) return;
    setBusy(true);
    clearMessage();
    try {
      await requestJSON(`/devices/${deviceId}`, {
        method: "DELETE",
        headers: authHeaders()
      });
      await loadDevices(true);
      showMessage("设备已删除", 1200);
    } catch (err) {
      showMessage(`删除设备失败: ${err.message}`);
    } finally {
      setBusy(false);
    }
  }

  function updateHistoryFilter(key, value) {
    setHistoryFilters((prev) => ({ ...prev, [key]: value }));
  }

  function resetHistoryFilters() {
    setHistoryFilters({
      from: "",
      to: "",
      parkId: "",
      treeId: "",
      deviceCode: ""
    });
  }

  function selectedParkName() {
    return parks.find((park) => park.id === selectedParkId)?.name || "";
  }

  function renderUploadPage() {
    return (
      <section className="panel content-panel">
        <div className="section-head">
          <div><h2>发送音频</h2></div>
          <span className="token-line">
            {referenceBusy ? "基础数据加载中..." : `令牌：${token.slice(0, 24)}...`}
          </span>
        </div>
        {!parks.length || !trees.length || !devices.length ? (
          <div className="empty">请先由管理员维护园区、树木和设备数据，再进行上传。</div>
        ) : (
          <form onSubmit={handleUpload} className="form">
            <label className="field">
              <span>园区</span>
              <select value={selectedParkId} onChange={(e) => setSelectedParkId(e.target.value)}>
                {parks.map((park) => (
                  <option key={park.id} value={park.id}>{park.name}</option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>树木编号</span>
              <select value={selectedTreeId} onChange={(e) => setSelectedTreeId(e.target.value)}>
                {treesForSelectedPark.map((tree) => (
                  <option key={tree.id} value={tree.id}>{tree.treeCode}</option>
                ))}
              </select>
            </label>
            <label className="field">
              <span>设备</span>
              <select value={selectedDeviceCode} onChange={(e) => setSelectedDeviceCode(e.target.value)}>
                {devicesForSelectedTree.map((device) => (
                  <option key={device.id} value={device.deviceCode}>
                    {device.deviceName} ({device.deviceCode})
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
        )}
      </section>
    );
  }

  function renderHistoryPage() {
    return (
      <section className="panel content-panel">
        <div className="section-head">
          <div><h2>历史检测记录</h2></div>
          <button className="ghost" disabled={jobsBusy} type="button" onClick={() => loadJobs()}>
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
            <select value={historyFilters.deviceCode} onChange={(e) => updateHistoryFilter("deviceCode", e.target.value)}>
              <option value="">全部设备</option>
              {devicesForHistoryFilter.map((device) => (
                <option key={device.id} value={device.deviceCode}>
                  {device.deviceName} ({device.deviceCode})
                </option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>园区</span>
            <select value={historyFilters.parkId} onChange={(e) => updateHistoryFilter("parkId", e.target.value)}>
              <option value="">全部园区</option>
              {parks.map((park) => (
                <option key={park.id} value={park.id}>{park.name}</option>
              ))}
            </select>
          </label>
          <label className="field">
            <span>树木编号</span>
            <select value={historyFilters.treeId} onChange={(e) => updateHistoryFilter("treeId", e.target.value)}>
              <option value="">全部树木</option>
              {treesForHistoryFilter.map((tree) => (
                <option key={tree.id} value={tree.id}>{tree.treeCode}</option>
              ))}
            </select>
          </label>
        </div>
        <div className="filter-actions">
          <button className="ghost" type="button" onClick={resetHistoryFilters}>清空筛选</button>
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
                    <p className="meta">
                      园区: {job.parkName || "未设置"} | 树木编号: {job.treeCode || "未设置"} | 设备: {job.deviceName || job.deviceCode || "未设置"}
                    </p>
                    <p className="meta">
                      任务ID: {job.jobId} | 时间: {parseDate(job.createdAt)?.toLocaleString() || job.createdAt}
                    </p>
                    {job.resultScore != null && <p className="meta">置信度: {job.resultScore.toFixed(6)}</p>}
                    {job.errorMessage && <p className="meta">错误信息: {job.errorMessage}</p>}
                  </div>
                  <div className={`status status-${getDetectionLabel(job)}`}>{getDetectionLabel(job)}</div>
                </div>
                <div className="job-actions">
                  <button className="danger" disabled={busy} type="button" onClick={() => handleDeleteJob(job.jobId)}>
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
        <div className="section-head">
          <div><h2>园区管理</h2></div>
          <button className="ghost" type="button" onClick={() => loadParks()}>刷新园区</button>
        </div>
        <div className="manage-grid">
          <form className="form mini-form" onSubmit={handleCreatePark}>
            <h3>新增园区</h3>
            <input placeholder="园区名称" value={parkForm.name} onChange={(e) => setParkForm((prev) => ({ ...prev, name: e.target.value }))} />
            <input placeholder="园区编码" value={parkForm.code} onChange={(e) => setParkForm((prev) => ({ ...prev, code: e.target.value }))} />
            <input placeholder="位置描述" value={parkForm.location} onChange={(e) => setParkForm((prev) => ({ ...prev, location: e.target.value }))} />
            <button className="ghost" type="submit" disabled={busy}>添加园区</button>
          </form>
          <div className="job-list">
            {parks.map((park) => (
              <article key={park.id} className="job-card">
                <div className="job-main">
                  <div>
                    <h3>{park.name}</h3>
                    <p className="meta">编码: {park.code}</p>
                    <p className="meta">位置: {park.location || "未填写"}</p>
                  </div>
                  <div className={`status status-${park.status === 1 ? "健康" : "检测失败"}`}>{park.status === 1 ? "正常" : "停用"}</div>
                </div>
                <div className="job-actions">
                  <button className="danger" type="button" disabled={busy} onClick={() => handleDeletePark(park.id)}>删除园区</button>
                </div>
              </article>
            ))}
          </div>
        </div>
      </section>
    );
  }

  function renderTreesPage() {
    return (
      <section className="panel content-panel">
        <div className="section-head">
          <div><h2>树木管理</h2></div>
          <button className="ghost" type="button" onClick={() => loadTrees()}>刷新树木</button>
        </div>
        <div className="manage-grid">
          <form className="form mini-form" onSubmit={handleCreateTree}>
            <h3>新增树木</h3>
            <select value={treeForm.parkId} onChange={(e) => setTreeForm((prev) => ({ ...prev, parkId: e.target.value }))}>
              {parks.map((park) => <option key={park.id} value={park.id}>{park.name}</option>)}
            </select>
            <input placeholder="树木编号" value={treeForm.treeCode} onChange={(e) => setTreeForm((prev) => ({ ...prev, treeCode: e.target.value }))} />
            <input placeholder="树种" value={treeForm.species} onChange={(e) => setTreeForm((prev) => ({ ...prev, species: e.target.value }))} />
            <button className="ghost" type="submit" disabled={busy}>添加树木</button>
          </form>
          <div className="job-list">
            {trees.map((tree) => (
              <article key={tree.id} className="job-card">
                <div className="job-main">
                  <div>
                    <h3>{tree.treeCode}</h3>
                    <p className="meta">园区: {tree.parkName || "未设置"}</p>
                    <p className="meta">树种: {tree.species || "未填写"}</p>
                  </div>
                  <div className={`status status-${tree.status === 1 ? "健康" : "检测失败"}`}>{tree.status === 1 ? "正常" : "停用"}</div>
                </div>
                <div className="job-actions">
                  <button className="danger" type="button" disabled={busy} onClick={() => handleDeleteTree(tree.id)}>删除树木</button>
                </div>
              </article>
            ))}
          </div>
        </div>
      </section>
    );
  }

  function renderDevicesPage() {
    return (
      <section className="panel content-panel">
        <div className="section-head">
          <div><h2>设备管理</h2></div>
          <button className="ghost" type="button" onClick={() => loadDevices()}>刷新设备</button>
        </div>
        <div className="manage-grid">
          <form className="form mini-form" onSubmit={handleCreateDevice}>
            <h3>新增设备</h3>
            <select value={deviceForm.parkId} onChange={(e) => setDeviceForm((prev) => ({ ...prev, parkId: e.target.value, treeId: "" }))}>
              {parks.map((park) => <option key={park.id} value={park.id}>{park.name}</option>)}
            </select>
            <select value={deviceForm.treeId} onChange={(e) => setDeviceForm((prev) => ({ ...prev, treeId: e.target.value }))}>
              {treesForDeviceForm.map((tree) => <option key={tree.id} value={tree.id}>{tree.treeCode}</option>)}
            </select>
            <input placeholder="设备编码，例如 dev_001" value={deviceForm.deviceCode} onChange={(e) => setDeviceForm((prev) => ({ ...prev, deviceCode: e.target.value }))} />
            <input placeholder="设备名称" value={deviceForm.name} onChange={(e) => setDeviceForm((prev) => ({ ...prev, name: e.target.value }))} />
            <button className="ghost" type="submit" disabled={busy}>添加设备</button>
          </form>
          <div className="job-list">
            {devices.map((device) => (
              <article key={device.id} className="job-card">
                <div className="job-main">
                  <div>
                    <h3>{device.deviceName}</h3>
                    <p className="meta">设备编码: {device.deviceCode}</p>
                    <p className="meta">园区: {device.parkName || "未设置"} | 树木: {device.treeCode || "未设置"}</p>
                  </div>
                  <div className={`status status-${device.status === 1 ? "健康" : "检测失败"}`}>{device.status === 1 ? "正常" : "停用"}</div>
                </div>
                <div className="job-actions">
                  <button className="danger" type="button" disabled={busy} onClick={() => handleDeleteDevice(device.id)}>删除设备</button>
                </div>
              </article>
            ))}
          </div>
        </div>
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

  const navItems = isAdmin
    ? [
        ["upload", "发送音频"],
        ["history", "历史记录"],
        ["parks", "园区管理"],
        ["trees", "树木管理"],
        ["devices", "设备管理"]
      ]
    : [
        ["upload", "发送音频"],
        ["history", "历史记录"]
      ];

  return (
    <div className="page">
      <div className="shell">
        <section className={`hero ${isAuthed ? "hero-bar" : "hero-center"}`}>
          <div className="hero-copy">
            <h1>病虫害监测系统</h1>
          </div>
          {isAuthed && (
            <button
              className="ghost"
              type="button"
              onClick={() => {
                setToken("");
                setUser(null);
                setPage("upload");
                showMessage("已退出登录", 1000);
              }}
            >
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
              <div className="pill-list" style={{ marginBottom: 16 }}>
                <span className="pill">{user?.username || "当前用户"}</span>
                <span className="pill">{isAdmin ? "管理员" : "普通用户"}</span>
                {selectedParkName() && <span className="pill">{selectedParkName()}</span>}
              </div>
              <div className="nav-list">
                {navItems.map(([key, label]) => (
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
