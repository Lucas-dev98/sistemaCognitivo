const state = {
  tasks: [],
  status: null,
  events: [],
};

const el = {
  waStatus: document.getElementById("waStatus"),
  groupStatus: document.getElementById("groupStatus"),
  modeStatus: document.getElementById("modeStatus"),
  dbStatus: document.getElementById("dbStatus"),
  persistStatus: document.getElementById("persistStatus"),
  lastUpdate: document.getElementById("lastUpdate"),
  metricTotal: document.getElementById("metricTotal"),
  metricPending: document.getElementById("metricPending"),
  metricSoon: document.getElementById("metricSoon"),
  metricReminded: document.getElementById("metricReminded"),
  taskCount: document.getElementById("taskCount"),
  taskList: document.getElementById("taskList"),
  eventList: document.getElementById("eventList"),
  ingestForm: document.getElementById("ingestForm"),
  message: document.getElementById("message"),
  feedback: document.getElementById("formFeedback"),
  refreshBtn: document.getElementById("refreshBtn"),
  semanticState: document.getElementById("semanticState"),
  semanticResult: document.getElementById("semanticResult"),
  semanticNotes: document.getElementById("semanticNotes"),
};

let semanticTimer = null;
let semanticRequestId = 0;

function addEvent(text) {
  const timestamp = new Date().toLocaleTimeString("pt-BR");
  state.events.unshift({ text, timestamp });
  state.events = state.events.slice(0, 12);

  el.eventList.innerHTML = "";
  state.events.forEach((entry) => {
    const li = document.createElement("li");
    li.className = "event-item";
    li.textContent = `[${entry.timestamp}] ${entry.text}`;
    el.eventList.appendChild(li);
  });
}

function parseDate(value) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

function renderStatus() {
  if (!state.status) {
    return;
  }

  el.waStatus.textContent = state.status.connected ? "Conectado" : "Desconectado";
  el.groupStatus.textContent = state.status.group_found ? "Vinculado" : "Não encontrado";
  el.modeStatus.textContent = state.status.mode || "-";
  const backend = state.status.store_backend || "desconhecido";
  const healthy = state.status.store_healthy;
  const persisted = state.status.store_persisted;

  el.dbStatus.textContent = healthy === false ? `${backend} (indisponível)` : backend;
  el.persistStatus.textContent = persisted ? "Ativa" : "Fallback memória";
  el.lastUpdate.textContent = new Date().toLocaleTimeString("pt-BR");
}

function renderMetrics() {
  const now = new Date();
  let pending = 0;
  let soon = 0;
  let reminded = 0;

  state.tasks.forEach((task) => {
    if (task.reminded) {
      reminded += 1;
      return;
    }

    pending += 1;
    const dueAt = parseDate(task.due_at);
    if (dueAt) {
      const diff = dueAt.getTime() - now.getTime();
      if (diff >= 0 && diff <= 24 * 60 * 60 * 1000) {
        soon += 1;
      }
    }
  });

  el.metricTotal.textContent = String(state.tasks.length);
  el.metricPending.textContent = String(pending);
  el.metricSoon.textContent = String(soon);
  el.metricReminded.textContent = String(reminded);
}

function renderTasks() {
  const tasks = [...state.tasks].sort((a, b) => new Date(a.due_at) - new Date(b.due_at));
  el.taskCount.textContent = `${tasks.length} itens`;
  el.taskList.innerHTML = "";

  if (tasks.length === 0) {
    const empty = document.createElement("li");
    empty.className = "task-item";
    empty.textContent = "Nenhuma tarefa cadastrada ainda.";
    el.taskList.appendChild(empty);
    return;
  }

  const now = Date.now();

  tasks.forEach((task) => {
    const li = document.createElement("li");
    li.className = "task-item";

    const dueAt = parseDate(task.due_at);
    const dueLabel = dueAt ? dueAt.toLocaleString("pt-BR") : task.due_at;

    const title = document.createElement("p");
    title.className = "title";
    title.textContent = task.title;

    const meta = document.createElement("p");
    meta.className = "meta";
    meta.textContent = `#${task.id} | prazo: ${dueLabel} | origem: ${task.source}`;

    const badge = document.createElement("span");
    badge.className = "badge";

    if (task.reminded) {
      badge.classList.add("done");
      badge.textContent = "Lembrete enviado";
    } else if (dueAt && dueAt.getTime() - now <= 30 * 60 * 1000) {
      badge.classList.add("soon");
      badge.textContent = "Janela de lembrete";
    } else {
      badge.classList.add("wait");
      badge.textContent = "Aguardando";
    }

    li.appendChild(title);
    li.appendChild(meta);
    li.appendChild(badge);
    el.taskList.appendChild(li);
  });
}

function renderAll() {
  renderStatus();
  renderMetrics();
  renderTasks();
}

function renderSemanticPreview(payload, sourceText) {
  el.semanticNotes.innerHTML = "";

  if (!sourceText) {
    el.semanticState.textContent = "aguardando texto";
    el.semanticResult.className = "semantic-result";
    el.semanticResult.textContent = "Digite uma mensagem para ver a leitura semântica em tempo real.";
    return;
  }

  if (!payload) {
    el.semanticState.textContent = "analisando...";
    el.semanticResult.className = "semantic-result";
    el.semanticResult.textContent = "Consultando o serviço semântico.";
    return;
  }

  if (payload.accepted && payload.task) {
    el.semanticState.textContent = "aceita";
    el.semanticResult.className = "semantic-result good";
    el.semanticResult.innerHTML = `<strong>Compromisso identificado.</strong><br>"${payload.task.title}"<br>Prazo sugerido: ${new Date(payload.task.due_at).toLocaleString("pt-BR")}`;
    addSemanticNote("A mensagem tem sinais fortes de compromisso e prazo.");
    addSemanticNote("Você pode enviar assim ou ajustar o contexto se quiser mais precisão.");
    return;
  }

  if (payload.needs_context) {
    el.semanticState.textContent = "precisa de contexto";
    el.semanticResult.className = "semantic-result warn";
    el.semanticResult.textContent = payload.error || "A mensagem parece compromisso, mas falta assunto ou local.";
    addSemanticNote("Exemplo melhor: Reunião com equipe comercial na sala 2 hoje 15h.");
    addSemanticNote("Inclua com quem, onde ou para qual assunto isso acontece.");
    return;
  }

  el.semanticState.textContent = "rejeitada";
  el.semanticResult.className = "semantic-result bad";
  el.semanticResult.textContent = payload.error || "A mensagem não parece compromisso.";
  addSemanticNote("O serviço semântico não encontrou intenção de tarefa suficiente.");
  addSemanticNote("Tente inserir verbo de ação + data/horário + contexto.");
}

function addSemanticNote(text) {
  const li = document.createElement("li");
  li.textContent = text;
  el.semanticNotes.appendChild(li);
}

async function fetchJSON(path) {
  const response = await fetch(path, {
    headers: {
      Accept: "application/json",
    },
  });

  if (!response.ok) {
    let message = `Erro ${response.status}`;
    try {
      const data = await response.json();
      if (data.error) {
        message = data.error;
      }
    } catch (_) {
      // ignore parse error
    }
    throw new Error(message);
  }

  return response.json();
}

async function refresh() {
  try {
    const [status, tasks] = await Promise.all([
      fetchJSON("/status"),
      fetchJSON("/tasks"),
    ]);

    state.status = status;
    state.tasks = tasks;
    renderAll();
  } catch (error) {
    addEvent(`Falha ao atualizar painel: ${error.message}`);
  }
}

async function createTask(message) {
  const response = await fetch("/ingest/whatsapp", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: JSON.stringify({ message }),
  });

  const data = await response.json().catch(() => ({}));

  if (!response.ok) {
    throw new Error(data.error || `Erro ${response.status}`);
  }

  return data;
}

async function analyzeSemantic(message) {
  const response = await fetch("/semantic/analyze", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
    },
    body: JSON.stringify({ message }),
  });

  const data = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(data.error || `Erro ${response.status}`);
  }

  return data;
}

function scheduleSemanticAnalysis() {
  const message = el.message.value.trim();
  semanticRequestId += 1;
  const currentRequestId = semanticRequestId;

  clearTimeout(semanticTimer);

  if (!message) {
    renderSemanticPreview(null, "");
    return;
  }

  renderSemanticPreview(null, message);

  semanticTimer = setTimeout(async () => {
    try {
      const payload = await analyzeSemantic(message);
      if (currentRequestId !== semanticRequestId) {
        return;
      }
      renderSemanticPreview(payload, message);
    } catch (error) {
      if (currentRequestId !== semanticRequestId) {
        return;
      }
      el.semanticState.textContent = "serviço indisponível";
      el.semanticResult.className = "semantic-result bad";
      el.semanticResult.textContent = error.message;
      el.semanticNotes.innerHTML = "";
      addSemanticNote("Não foi possível consultar o serviço semântico agora.");
      addSemanticNote("O envio da tarefa continua funcionando pelo fluxo normal.");
    }
  }, 350);
}

el.ingestForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const message = el.message.value.trim();
  if (!message) {
    return;
  }

  el.feedback.textContent = "Enviando...";
  el.feedback.className = "feedback";

  try {
    const task = await createTask(message);
    addEvent(`Tarefa #${task.id} criada: ${task.title}`);
    el.feedback.textContent = "Tarefa criada com sucesso.";
    el.feedback.className = "feedback ok";
    el.message.value = "";
    await refresh();
  } catch (error) {
    el.feedback.textContent = error.message;
    el.feedback.className = "feedback error";
    addEvent(`Erro ao criar tarefa: ${error.message}`);
  }
});

el.message.addEventListener("input", scheduleSemanticAnalysis);

el.refreshBtn.addEventListener("click", async () => {
  addEvent("Atualização manual executada");
  await refresh();
});

addEvent("Painel iniciado");
renderSemanticPreview(null, "");
refresh();
setInterval(refresh, 7000);
