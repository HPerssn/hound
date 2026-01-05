let es;
let sessionId;
let sessionData;

function formatTime(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
}

function parseTimeInput(timeStr) {
    if (!timeStr) return 0;

    // Remove all whitespace
    timeStr = timeStr.trim();

    // Handle mm:ss format
    if (timeStr.includes(':')) {
        const parts = timeStr.split(':');
        const mins = parseInt(parts[0]) || 0;
        const secs = parseInt(parts[1]) || 0;
        return mins * 60 + secs;
    }

    // Handle plain numbers - if 3 or more digits, treat as mm:ss without colon
    // e.g., "503" becomes 5:03, "130" becomes 1:30
    if (timeStr.length >= 3) {
        const mins = parseInt(timeStr.slice(0, -2)) || 0;
        const secs = parseInt(timeStr.slice(-2)) || 0;
        return mins * 60 + secs;
    }

    // Handle 1-2 digit numbers as minutes
    const num = parseInt(timeStr);
    return isNaN(num) ? 0 : num * 60;
}

async function startSession() {
    const timeInput = document.getElementById("targetMin").value;
    const targetSec = parseTimeInput(timeInput);

    if (!targetSec || targetSec <= 0) {
        alert("Please enter a valid target time (e.g., 05:00 or 5:00)");
        return;
    }

    try {
        const res = await fetch('/sessions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ targetSec })
        });

        if (!res.ok) {
            throw new Error("Failed to start session");
        }

        sessionData = await res.json();
        sessionId = sessionData.ID;

        if (sessionData.Steps && sessionData.Steps.length > 0) {
            displaySteps(sessionData.Steps);
        }

        connectToSession();
        document.getElementById("activeStep").textContent = `Session ${sessionId} started`;
    } catch (err) {
        console.error("Error in startSession:", err);
        alert("Error starting session: " + err.message);
    }
}

function displaySteps(steps) {
    const container = document.getElementById("steps");
    container.innerHTML = "";

    steps.forEach(step => {
        const div = document.createElement("div");
        div.className = "step";
        div.id = `step-${step.Index}`;
        div.innerHTML = `
            <span class="step-label">Step ${step.Index + 1} - ${formatTime(step.Duration)}</span>
	    <span class="step-timer" id="timer-${step.Index}">00:00</span>
            <div class="step-actions">
                <button onclick="startStep(${step.Index})">Start</button>
                <button onclick="stopStep(${step.Index})">Stop</button>
            </div>
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
            if (timerEl) timerEl.textContent = formatTime(data.elapsed);
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

async function completeSession(successLevel) {
    if (!sessionId) return;

    const comment = prompt("Optional comment for this session:", "");

    try {
        const res = await fetch(`/sessions/${sessionId}/complete`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                success: successLevel,
                comment: comment || ""
            })
        });

        if (!res.ok) {
            const err = await res.json();
            throw new Error(err.error || "Failed to save session");
        }

        document.getElementById("activeStep").textContent = "Session saved";
        if (es) es.close();
        sessionId = null;

    } catch (err) {
        console.error(err);
        alert(err.message);
    }
}
