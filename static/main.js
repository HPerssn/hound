let es;
let sessionId;
let sessionData;

async function startSession() {
    const targetMin = parseInt(document.getElementById("targetMin").value, 10);
    if (!targetMin) {
        alert("Please enter target time in minutes");
    }
    const targetSec = targetMin * 60;

    try {
        const res = await fetch('/sessions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ targetSec })
        });

        if (!res.ok) throw new Error("failed to start session");

        sessionData = await res.json();
        sessionId = sessionData.ID;

        displaySteps(sessionData.Steps);
        connectToSession();
        document.getElementById("activeStep").textContent = `Session ${sessionId} started`;
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

    };

    es.onerror = () => {
        console.log("EventSource connection closed");
        es.close();
    }
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
    try {
        const res = await fetch(`/sessions/${sessionId}/stop`, { method: 'POST' });
        if (!res.ok) {
            const err = await res.json();
            throw new Error(err.error || "Failed to stop session");
        }

        document.getElementById("activeStep").textContent = "Session stopped";
        if (es) es.close();
    } catch (err) {
        console.error(err);
        alert(err.message);
    }
}

