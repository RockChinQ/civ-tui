// civ-tui web client. Single-player local UI: char-style map + side panel.

const TERRAIN_CLASS = {
  "Ocean": "t-ocean",
  "Coast": "t-coast",
  "Grassland": "t-grass",
  "Plains": "t-plains",
  "Hills": "t-hills",
  "Mountains": "t-mtn",
  "Forest": "t-forest",
  "Desert": "t-desert",
  "Tundra": "t-tundra",
};

const state = {
  game: null,
  cursor: { x: 0, y: 0 },
  selectedUnitId: null,
  rangeMode: false,
  destMode: false,
};

// --- API helpers --------------------------------------------------------

async function api(path, body) {
  const opts = { method: body !== undefined ? "POST" : "GET" };
  if (body !== undefined) {
    opts.headers = { "Content-Type": "application/json" };
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(path, opts);
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    addLocalLog("Error: " + (err.error || res.status));
    return null;
  }
  return res.json();
}

async function refreshState() {
  const s = await api("/api/state");
  if (s && s.state && s.state !== "none") {
    state.game = s;
  } else {
    state.game = null;
  }
  render();
}

async function applyState(promise) {
  const s = await promise;
  if (s && s.turn !== undefined) {
    state.game = s;
    render();
  }
}

// --- Rendering ----------------------------------------------------------

function render() {
  const s = state.game;
  const status = document.getElementById("status");
  if (!s) {
    status.textContent = "No active game. Click New.";
    document.getElementById("map").innerHTML = "";
    document.getElementById("tile-info").innerHTML = "";
    document.getElementById("unit-detail").innerHTML = "";
    document.getElementById("civ-detail").innerHTML = "";
    document.getElementById("messages").innerHTML = "";
    document.getElementById("unit-actions").innerHTML = "";
    return;
  }

  const player = s.civs.find(c => c.isPlayer) || {};
  const stateLabel = s.state === "running" ? "" :
    s.state === "victory" ? " — VICTORY" :
    s.state === "defeat" ? " — DEFEAT" :
    s.state === "draw" ? " — DRAW" : "";
  status.textContent = `Turn ${s.turn}/${s.maxTurns} · ${player.name || "?"} · Gold ${player.gold} · Sci ${player.science}${stateLabel}`;

  renderMap();
  renderTileInfo();
  renderSelection();
  renderCiv();
  renderMessages();
}

function renderMap() {
  const s = state.game;
  const map = document.getElementById("map");

  // Build unit/city lookup by tile
  const unitAt = new Map();
  for (const u of s.units) unitAt.set(u.y * 1000 + u.x, u);
  const cityAt = new Map();
  for (const c of s.cities) cityAt.set(c.y * 1000 + c.x, c);

  const sel = state.selectedUnitId != null
    ? s.units.find(u => u.id === state.selectedUnitId) : null;

  const reachable = new Set();
  if (sel && sel.movesLeft > 0) {
    // Local approximation of reachable: 8 neighbors, terrain-passable, cost <= moves
    const dirs = [[-1,-1],[0,-1],[1,-1],[-1,0],[1,0],[-1,1],[0,1],[1,1]];
    for (const [dx, dy] of dirs) {
      const nx = sel.x + dx, ny = sel.y + dy;
      if (nx < 0 || ny < 0 || nx >= s.mapWidth || ny >= s.mapHeight) continue;
      const t = s.tiles[ny][nx];
      if (!t.passable) continue;
      if (t.moveCost > sel.movesLeft) continue;
      reachable.add(ny * 1000 + nx);
    }
  }

  // Range targets
  const rangeTargets = new Set();
  if (state.rangeMode && sel && sel.range > 0) {
    for (let dy = -sel.range; dy <= sel.range; dy++) {
      for (let dx = -sel.range; dx <= sel.range; dx++) {
        if (Math.abs(dx) + Math.abs(dy) > sel.range) continue;
        const nx = sel.x + dx, ny = sel.y + dy;
        if (nx < 0 || ny < 0 || nx >= s.mapWidth || ny >= s.mapHeight) continue;
        rangeTargets.add(ny * 1000 + nx);
      }
    }
  }

  // Build innerHTML in one pass — much faster than DOM ops per cell
  const rows = [];
  for (let y = 0; y < s.mapHeight; y++) {
    const cells = [];
    for (let x = 0; x < s.mapWidth; x++) {
      const t = s.tiles[y][x];
      const key = y * 1000 + x;
      let cls = "cell";
      let symbol = " ";
      let civClass = "";

      if (!t.revealed) {
        cls += " unrevealed";
        symbol = " ";
      } else {
        cls += " " + (TERRAIN_CLASS[t.terrain] || "t-grass");
        symbol = t.symbol || ".";
        if (t.impSymbol) symbol = t.impSymbol;
        if (!t.visible) cls += " fog";

        const c = cityAt.get(key);
        const u = unitAt.get(key);
        if (c) {
          cls += " has-city civ-" + c.civId;
          symbol = c.name[0] || "C";
        } else if (u && (t.visible || u.civId === 1)) {
          cls += " has-unit civ-" + u.civId;
          symbol = u.symbol;
        }
      }

      if (reachable.has(key)) cls += " reachable";
      if (rangeTargets.has(key)) cls += " range-target";
      if (sel && sel.x === x && sel.y === y) cls += " selected";
      if (state.cursor.x === x && state.cursor.y === y) cls += " cursor";

      // Escape special HTML chars
      const safe = symbol === "<" ? "&lt;" : symbol === ">" ? "&gt;" : symbol === "&" ? "&amp;" : symbol;
      cells.push(`<span class="${cls}" data-x="${x}" data-y="${y}">${safe}</span>`);
    }
    rows.push(`<div class="row">${cells.join("")}</div>`);
  }
  map.innerHTML = rows.join("");
}

function renderTileInfo() {
  const s = state.game;
  const { x, y } = state.cursor;
  if (x < 0 || y < 0 || x >= s.mapWidth || y >= s.mapHeight) return;
  const t = s.tiles[y][x];
  const div = document.getElementById("tile-info");
  if (!t.revealed) {
    div.innerHTML = `<div class="kv"><span>(${x},${y})</span><span>?</span></div>`;
    return;
  }
  let html = `<div class="kv"><span>Pos</span><span>(${x},${y})</span></div>`;
  html += `<div class="kv"><span>Terrain</span><span>${t.terrain}</span></div>`;
  html += `<div class="kv"><span>Yield</span><span>F${t.food} P${t.prod} G${t.gold}</span></div>`;
  html += `<div class="kv"><span>Move cost</span><span>${t.moveCost}</span></div>`;
  if (t.imp) html += `<div class="kv"><span>Improvement</span><span>${t.imp}</span></div>`;
  if (t.defBonus > 0) html += `<div class="kv"><span>Defense</span><span>+${t.defBonus}%</span></div>`;
  div.innerHTML = html;
}

function renderSelection() {
  const s = state.game;
  const { x, y } = state.cursor;
  const detail = document.getElementById("unit-detail");
  const actions = document.getElementById("unit-actions");

  const u = s.units.find(u => u.x === x && u.y === y);
  const c = s.cities.find(c => c.x === x && c.y === y);

  let html = "";
  let buttons = [];

  // Auto-select friendly unit when cursor lands on it
  if (u && u.civId === 1) {
    state.selectedUnitId = u.id;
  } else if (!u) {
    // keep selectedUnitId so movement works from sidebar; but show nothing if none on cursor
  }

  if (c) {
    html += `<div><b>${c.name}</b> (Civ ${c.civId})</div>`;
    html += `<div class="kv"><span>Pop</span><span>${c.pop}</span></div>`;
    html += `<div class="kv"><span>Food</span><span>${c.food}/${c.foodNeeded}</span></div>`;
    html += `<div class="kv"><span>Prod</span><span>${c.prod}</span></div>`;
    html += `<div class="kv"><span>HP</span><span>${c.hp}/${c.maxHp}</span></div>`;
    if (c.queue && c.queue.length) {
      html += `<div style="margin-top:4px;">Queue:</div>`;
      for (const q of c.queue) {
        html += `<div class="kv"><span>${q.name}</span><span>${q.cost}</span></div>`;
      }
    }
    if (c.buildings && c.buildings.length) {
      html += `<div style="margin-top:4px;">Buildings: ${c.buildings.join(", ")}</div>`;
    }
    if (c.civId === 1) {
      buttons.push(`<button data-act="build-city" data-id="${c.id}">Build…</button>`);
    }
  }

  if (u) {
    if (c) html += `<hr/>`;
    html += `<div><b>${u.type}</b> (Civ ${u.civId})</div>`;
    html += `<div class="kv"><span>HP</span><span>${u.hp}/${u.maxHp}</span></div>`;
    html += `<div class="kv"><span>ATK/DEF</span><span>${u.attack}/${u.defense}</span></div>`;
    html += `<div class="kv"><span>Moves</span><span>${u.movesLeft}/${u.maxMoves}</span></div>`;
    if (u.range > 0) html += `<div class="kv"><span>Range</span><span>${u.range}</span></div>`;
    if (u.building) html += `<div class="kv"><span>Building</span><span>${u.building} (${u.buildLeft}t)</span></div>`;
    if (u.hasDest) html += `<div class="kv"><span>Dest</span><span>(${u.destX},${u.destY})</span></div>`;
    if (u.civId === 1 && u.movesLeft > 0) {
      if (u.type === "Settler") buttons.push(`<button data-act="found">Found City (F)</button>`);
      if (u.type === "Worker") buttons.push(`<button data-act="improve">Improve (I)</button>`);
      if (u.range > 0) buttons.push(`<button data-act="range">Ranged (R)</button>`);
      buttons.push(`<button data-act="dest">Set Dest (G)</button>`);
      buttons.push(`<button data-act="wait">Wait (Z)</button>`);
      if (u.hasDest) buttons.push(`<button data-act="cleardest">Clear Dest</button>`);
    }
  }

  if (!u && !c) html = `<em>(empty)</em>`;
  detail.innerHTML = html;
  actions.innerHTML = buttons.join("");
}

function renderCiv() {
  const s = state.game;
  const player = s.civs.find(c => c.isPlayer);
  if (!player) return;
  let html = `<div><b>${player.name}</b></div>`;
  html += `<div class="kv"><span>Gold</span><span>${player.gold}</span></div>`;
  html += `<div class="kv"><span>Science</span><span>${player.science}</span></div>`;
  if (player.researching) {
    html += `<div class="kv"><span>Researching</span><span>${player.researching} (${player.progress})</span></div>`;
  } else {
    html += `<div class="kv"><span>Researching</span><span><em>none</em></span></div>`;
  }
  html += `<div class="kv"><span>Techs</span><span>${(player.techs||[]).length}</span></div>`;
  // Other civs
  const others = s.civs.filter(c => !c.isPlayer);
  if (others.length) {
    html += `<div style="margin-top:4px;">Other civs:</div>`;
    for (const o of others) {
      const rel = (player.relations || {})[o.id] || "peace";
      const status = !o.isAlive ? "dead" : rel;
      html += `<div class="kv"><span>${o.name}</span><span>${status}</span></div>`;
    }
  }
  document.getElementById("civ-detail").innerHTML = html;
}

function renderMessages() {
  const s = state.game;
  const div = document.getElementById("messages");
  // newest first; show last 30
  const items = (s.messages || []).slice(-30).reverse();
  div.innerHTML = items.map(m => {
    const cls = m.isPlayer ? "msg player" : "msg";
    return `<div class="${cls}">${escapeHtml(m.text)}</div>`;
  }).join("");
}

function escapeHtml(s) {
  return String(s).replace(/[&<>]/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;"}[c]));
}

function addLocalLog(text) {
  if (!state.game) return;
  state.game.messages = state.game.messages || [];
  state.game.messages.push({ text, isPlayer: true });
  renderMessages();
}

// --- Input --------------------------------------------------------------

function moveCursor(dx, dy) {
  if (!state.game) return;
  const nx = Math.max(0, Math.min(state.game.mapWidth - 1, state.cursor.x + dx));
  const ny = Math.max(0, Math.min(state.game.mapHeight - 1, state.cursor.y + dy));
  state.cursor = { x: nx, y: ny };
  scrollCursorIntoView();
  render();
}

function scrollCursorIntoView() {
  const map = document.getElementById("map");
  const cell = map.querySelector(`.cell[data-x="${state.cursor.x}"][data-y="${state.cursor.y}"]`);
  if (cell) cell.scrollIntoView({ block: "nearest", inline: "nearest" });
}

async function tryMoveSelectedTo(x, y) {
  const sel = getSelectedUnit();
  if (!sel) return false;
  const dx = x - sel.x, dy = y - sel.y;
  if (dx === 0 && dy === 0) return false;
  if (Math.abs(dx) > 1 || Math.abs(dy) > 1) return false;
  await applyState(api("/api/move", { unitId: sel.id, dx, dy }));
  return true;
}

function getSelectedUnit() {
  if (!state.game || state.selectedUnitId == null) return null;
  return state.game.units.find(u => u.id === state.selectedUnitId) || null;
}

async function onCellClick(x, y) {
  state.cursor = { x, y };

  if (state.rangeMode) {
    const sel = getSelectedUnit();
    if (sel) {
      await applyState(api("/api/ranged", { unitId: sel.id, x, y }));
    }
    state.rangeMode = false;
    render();
    return;
  }
  if (state.destMode) {
    const sel = getSelectedUnit();
    if (sel) {
      await applyState(api("/api/set-dest", { unitId: sel.id, x, y }));
    }
    state.destMode = false;
    render();
    return;
  }

  // If clicking adjacent and have selected unit, attempt move
  const sel = getSelectedUnit();
  if (sel && (Math.abs(x - sel.x) <= 1 && Math.abs(y - sel.y) <= 1) && !(x === sel.x && y === sel.y)) {
    const moved = await tryMoveSelectedTo(x, y);
    if (moved) return;
  }

  // Otherwise just select whatever unit is at the tile
  const u = state.game.units.find(u => u.x === x && u.y === y);
  if (u && u.civId === 1) {
    state.selectedUnitId = u.id;
  } else if (!u) {
    // Keep selection if any, just update cursor
  }
  render();
}

function bindMapClicks() {
  document.getElementById("map").addEventListener("click", (e) => {
    const cell = e.target.closest(".cell");
    if (!cell) return;
    const x = parseInt(cell.dataset.x, 10);
    const y = parseInt(cell.dataset.y, 10);
    onCellClick(x, y);
  });
}

function bindKeys() {
  document.addEventListener("keydown", async (e) => {
    if (document.getElementById("modal").classList.contains("hidden") === false) {
      if (e.key === "Escape") closeModal();
      return;
    }
    if (!state.game) return;

    let handled = true;
    switch (e.key) {
      case "ArrowUp": case "w": case "k": moveCursor(0, -1); break;
      case "ArrowDown": case "s": case "j": moveCursor(0, 1); break;
      case "ArrowLeft": case "a": case "h": moveCursor(-1, 0); break;
      case "ArrowRight": case "d": case "l": moveCursor(1, 0); break;
      case "Enter": {
        const sel = getSelectedUnit();
        if (sel) await tryMoveSelectedTo(state.cursor.x, state.cursor.y);
        break;
      }
      case "f": case "F": await invokeAction("found"); break;
      case "b": case "B": openBuildMenu(); break;
      case "t": case "T": openTechMenu(); break;
      case "r": case "R": await invokeAction("range"); break;
      case "g": case "G": await invokeAction("dest"); break;
      case "z": case "Z": await invokeAction("wait"); break;
      case "i": case "I": await invokeAction("improve"); break;
      case " ": await applyState(api("/api/end-turn", {})); break;
      case "Escape":
        state.rangeMode = false; state.destMode = false;
        state.selectedUnitId = null;
        render();
        break;
      default: handled = false;
    }
    if (handled) e.preventDefault();
  });
}

function bindButtons() {
  document.getElementById("btn-end-turn").addEventListener("click",
    () => applyState(api("/api/end-turn", {})));
  document.getElementById("btn-build").addEventListener("click", openBuildMenu);
  document.getElementById("btn-tech").addEventListener("click", openTechMenu);
  document.getElementById("btn-save").addEventListener("click", openSaveMenu);
  document.getElementById("btn-load").addEventListener("click", openLoadMenu);
  document.getElementById("btn-new").addEventListener("click", openNewMenu);
  document.getElementById("modal-close").addEventListener("click", closeModal);

  document.getElementById("unit-actions").addEventListener("click", async (e) => {
    const btn = e.target.closest("button[data-act]");
    if (!btn) return;
    const act = btn.dataset.act;
    if (act === "build-city") {
      const id = parseInt(btn.dataset.id, 10);
      openBuildMenuForCity(id);
    } else {
      await invokeAction(act);
    }
  });
}

async function invokeAction(act) {
  const sel = getSelectedUnit();
  switch (act) {
    case "found":
      if (sel) await applyState(api("/api/found-city", { unitId: sel.id }));
      break;
    case "wait":
      if (sel) await applyState(api("/api/wait", { unitId: sel.id }));
      break;
    case "range":
      if (sel && sel.range > 0) {
        state.rangeMode = true;
        addLocalLog("Ranged: click target tile (Esc to cancel)");
        render();
      }
      break;
    case "dest":
      if (sel) {
        state.destMode = true;
        addLocalLog("Destination: click target tile (Esc to cancel)");
        render();
      }
      break;
    case "cleardest":
      if (sel) await applyState(api("/api/clear-dest", { unitId: sel.id }));
      break;
    case "improve":
      if (sel && sel.type === "Worker") openImproveMenu(sel);
      break;
  }
}

// --- Modal menus --------------------------------------------------------

function openModal(title, bodyHtml, handler) {
  document.getElementById("modal-title").textContent = title;
  document.getElementById("modal-body").innerHTML = bodyHtml;
  document.getElementById("modal").classList.remove("hidden");
  if (handler) {
    document.getElementById("modal-body").addEventListener("click", handler, { once: false });
  }
}
function closeModal() {
  document.getElementById("modal").classList.add("hidden");
  document.getElementById("modal-body").innerHTML = "";
}

function openBuildMenu() {
  // Find player city under cursor first; otherwise pick first player city
  const s = state.game;
  if (!s) return;
  const cur = s.cities.find(c => c.civId === 1 && c.x === state.cursor.x && c.y === state.cursor.y);
  const city = cur || s.cities.find(c => c.civId === 1);
  if (!city) { addLocalLog("No city to build in"); return; }
  openBuildMenuForCity(city.id);
}

function openBuildMenuForCity(cityId) {
  const s = state.game;
  const city = s.cities.find(c => c.id === cityId);
  if (!city) return;
  const opts = s.buildOptions || [];
  let body = `<div style="margin-bottom:6px;">City: <b>${city.name}</b></div>`;
  body += opts.map(o => `
    <div class="list-item" data-iu="${o.isUnit ? 1 : 0}" data-name="${escapeHtml(o.name)}">
      <span>${o.isUnit ? "⚔" : "■"} ${o.name}</span>
      <span class="meta">${o.cost}p</span>
    </div>`).join("");
  openModal("Build / Train", body, async (e) => {
    const item = e.target.closest(".list-item");
    if (!item) return;
    const isUnit = item.dataset.iu === "1";
    const name = item.dataset.name;
    closeModal();
    await applyState(api("/api/build", { cityId, isUnit, name }));
  });
}

function openTechMenu() {
  const s = state.game;
  if (!s) return;
  const opts = s.techOptions || [];
  if (!opts.length) { addLocalLog("No techs available"); return; }
  const body = opts.map(o => `
    <div class="list-item" data-name="${escapeHtml(o.name)}">
      <span>${o.name}</span>
      <span class="meta">${o.progress}/${o.cost}</span>
    </div>`).join("");
  openModal("Research", body, async (e) => {
    const item = e.target.closest(".list-item");
    if (!item) return;
    const name = item.dataset.name;
    closeModal();
    await applyState(api("/api/research", { name }));
  });
}

function openImproveMenu(unit) {
  const opts = ["Farm", "Mine", "Lumber Mill", "Road"];
  const body = opts.map(o => `
    <div class="list-item" data-imp="${o}">
      <span>${o}</span>
    </div>`).join("");
  openModal("Worker — Build Improvement", body, async (e) => {
    const item = e.target.closest(".list-item");
    if (!item) return;
    const type = item.dataset.imp;
    closeModal();
    await applyState(api("/api/improvement", { unitId: unit.id, type }));
  });
}

async function openSaveMenu() {
  const saves = await api("/api/saves");
  if (!saves) return;
  const body = saves.map(sv => `
    <div class="list-item" data-slot="${sv.slot}">
      <span>${sv.empty ? `Slot ${sv.slot} — empty` : sv.label}</span>
      <span class="meta">save</span>
    </div>`).join("");
  openModal("Save Game — choose slot", body, async (e) => {
    const item = e.target.closest(".list-item");
    if (!item) return;
    const slot = parseInt(item.dataset.slot, 10);
    const r = await api("/api/save", { slot });
    if (r && r.ok) addLocalLog(`Saved to slot ${slot}`);
    closeModal();
  });
}

async function openLoadMenu() {
  const saves = await api("/api/saves");
  if (!saves) return;
  const body = saves.map(sv => `
    <div class="list-item ${sv.empty ? "disabled" : ""}" data-slot="${sv.slot}" data-empty="${sv.empty}">
      <span>${sv.empty ? `Slot ${sv.slot} — empty` : sv.label}</span>
      <span class="meta">load</span>
    </div>`).join("");
  openModal("Load Game", body, async (e) => {
    const item = e.target.closest(".list-item");
    if (!item || item.dataset.empty === "true") return;
    const slot = parseInt(item.dataset.slot, 10);
    closeModal();
    await applyState(api("/api/load", { slot }));
  });
}

function openNewMenu() {
  const body = `
    <div style="display:flex; flex-direction:column; gap:6px;">
      <label>Number of AI civs:
        <select id="new-ai">
          <option value="1">1</option><option value="2">2</option>
          <option value="3" selected>3</option><option value="4">4</option>
        </select>
      </label>
      <label>Map size:
        <select id="new-size">
          <option value="0">Small</option>
          <option value="1" selected>Medium</option>
          <option value="2">Large</option>
        </select>
      </label>
      <label>Difficulty:
        <select id="new-diff">
          <option value="1" selected>Normal</option>
          <option value="2">Hard</option>
          <option value="3">Brutal</option>
        </select>
      </label>
      <button id="new-go">Start</button>
    </div>`;
  openModal("New Game", body);
  document.getElementById("new-go").addEventListener("click", async () => {
    const numAI = parseInt(document.getElementById("new-ai").value, 10);
    const mapSize = parseInt(document.getElementById("new-size").value, 10);
    const diff = parseInt(document.getElementById("new-diff").value, 10);
    closeModal();
    await applyState(api("/api/new", { numAI, mapSize, difficulty: diff }));
    // Center cursor on player settler
    const settler = state.game.units.find(u => u.civId === 1 && u.type === "Settler");
    if (settler) {
      state.cursor = { x: settler.x, y: settler.y };
      state.selectedUnitId = settler.id;
      render();
      scrollCursorIntoView();
    }
  });
}

// --- Boot ---------------------------------------------------------------

window.addEventListener("DOMContentLoaded", async () => {
  bindMapClicks();
  bindButtons();
  bindKeys();
  await refreshState();
  if (!state.game) {
    openNewMenu();
  }
});
