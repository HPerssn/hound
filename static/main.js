let es;
let sessionId;
let sessionData;

async function startSession() {
    const targetMin = parseInt(document.getElementById("targetMin").value, 10);
    if (!targetMin) return;
    const targetSec = targetMin * 60;

    try {
        const res = await fetch('/sessions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ targetSec }) // only targetSec, ID is generated server-side
        });

        if (!res.ok) throw new Error("failed to start session");

        sessionData = await res.json();
        sessionId = sessionData.ID;

        displaySteps(sessionData.Steps);
        connectToSession();
    } catch (err) {
        console.error(err);
    }
}

function displaySteps(steps) {
    const container = document.getElementById("steps");
    container.innerHTML = "";

    steps.forEach(step => {
        const div = document.createElement("div");
        div.id = `step-${step.Index}`;
        div.innerHTML = `
            <span>Step ${step.Index + 1} - ${step.Duration}s</span>
            <button onclick="startStep(${step.Index})">Start</button>
            <button onclick="stopStep(${step.Index})">Stop</button>
            <span id="timer-${step.Index}">0s</span>
        `;
        container.appendChild(div);
    });
}

function connectToSession() {
    if (!sessionId) return;
    if (es) es.close();

    es = new EventSource(`/sessions/${sessionId}/events`);

    es.onmessage = (e) => {
        const data = JSON.parse(e.data);

        if (data.index !== undefined && data.elapsed !== undefined) {
            const timerEl = document.getElementById(`timer-${data.index}`);
            if (timerEl) timerEl.textContent = `${data.elapsed}s`;
        }

        if (data.completed) {
            const timerEl = document.getElementById(`timer-${data.index}`);
            if (timerEl) timerEl.textContent = "Completed";
        }

        if (data.type === "done") {
            document.getElementById("activeStep").textContent = "Session finished";
        }
    };

    es.onerror = () => es.close();
}

async function startStep(idx) {
    if (!sessionId) return;
    try {
        const res = await fetch(`/sessions/${sessionId}/steps/${idx}/start`, { method: "POST" });
        if (!res.ok) throw new Error("failed to start step");
    } catch (err) {
        console.error(err);
    }
}

async function stopStep(idx) {
    if (!sessionId) return;
    try {
        const res = await fetch(`/sessions/${sessionId}/steps/${idx}/stop`, { method: "POST" });
        if (!res.ok) throw new Error("failed to stop step");
    } catch (err) {
        console.error(err);
    }
}

async function stopSession() {
    if (!sessionId) return;
    await fetch(`/sessions/${sessionId}/stop`, { method: 'POST' });
    document.getElementById("activeStep").textContent = "Session stopped";
}

