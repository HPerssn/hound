let es;
let sessionId;
let sessionData;
let notificationsEnabled = false;

if ("Notification" in window) {
    Notification.requestPermission().then(permission => {
        notificationsEnabled = permission === "granted";
    });
}

function setActiveSession(id) {
    sessionId = id;
    if (id) {
        sessionStorage.setItem('hound_activeSession', id);
    } else {
        sessionStorage.removeItem('hound_activeSession');
    }
}

function getActiveSession() {
    return sessionStorage.getItem('hound_activeSession');
}

document.addEventListener("DOMContentLoaded", () => {
    hydrateTargetInput();

    const activeSessionId = getActiveSession();
    if (activeSessionId) {
        sessionId = activeSessionId;
        reconnectToSession();
    }
})

function hydrateTargetInput() {
    fetch('/next-target')
        .then(res => res.json())
        .then(data => {
            if (!data || !data.nextTarget || data.nextTarget <= 0) return;
            document.getElementById("targetMin").value = formatTime(data.nextTarget);
        })
        .catch(err => {
            console.error("Failed to hydrate target input:", err);
        });
}


document.addEventListener("visibilitychange", () => {
    if (!document.hidden && sessionId) {
        console.log("page became visible, reconnecting")
        reconnectToSession();
    }
});

function notifyUser(title, body) {
    if (!notificationsEnabled || !("Notification" in window)) return;

    //notify on phone lock/tab inactive
    if (document.hidden) {
        new Notification(title, {
            body: body,
            icon: "static/icon.png",
            tag: "hound-timer"
        });
    }
}

function reconnectToSession() {
    if (!sessionId) return;

    if (es) {
        es.close();
    }

    fetch(`/sessions/${sessionId}`)
        .then(res => {
            if (!res.ok) {
                console.log("Session not found on server, clearing local reference");
                setActiveSession(null);
                document.getElementById("activeStep").textContent = "Session expired. Start a new session.";
                document.getElementById("steps").innerHTML = "";
                return null;
            }
            return res.json();
        })
        .then(data => {
            if (!data) return;

            sessionData = data;
            displaySteps(sessionData.Steps);
            connectToSession();

            const targetTime = formatTime(sessionData.TargetSec);
            if (sessionData.Completed) {
                document.getElementById("activeStep").textContent = `Session completed - target was: ${targetTime}`;
            } else {
                document.getElementById("activeStep").textContent = `Session reconnected - target: ${targetTime}`;
            }
        })
        .catch(err => {
            console.error("Failed to reconnect:", err);
            setActiveSession(null);
            document.getElementById("activeStep").textContent = "Connection failed. Start a new session.";
        })
}

function formatTime(seconds) {
    if (!seconds || isNaN(seconds)) {
        return "00:00"
    }
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
}

function parseTimeInput(timeStr) {
    if (!timeStr) return 0;


    timeStr = timeStr.trim();

    // Handle mm:ss format
    if (timeStr.includes(':')) {
        const parts = timeStr.split(':');
        const mins = parseInt(parts[0]) || 0;
        const secs = parseInt(parts[1]) || 0;
        return mins * 60 + secs;
    }


    // e.g., "503" becomes 5:03, "130" becomes 1:30
    if (timeStr.length >= 3) {
        const mins = parseInt(timeStr.slice(0, -2)) || 0;
        const secs = parseInt(timeStr.slice(-2)) || 0;
        return mins * 60 + secs;
    }


    const num = parseInt(timeStr);
    return isNaN(num) ? 0 : num * 60;
}

async function startSession() {
    const timeInput = document.getElementById("targetMin").value;

    let targetSec = null;

    if (timeInput && timeInput.trim() !== "") {
        targetSec = parseTimeInput(timeInput);
        if (!targetSec || targetSec <= 0) {
            alert("Please enter a valid target time (e.g., 05:00 or 5:00)");
            return;
        }
    }

    try {
        const body = {};
        if (targetSec) {
            body.targetSec = targetSec;
        }

        const res = await fetch('/sessions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body)
        });

        if (!res.ok) {
            throw new Error("Failed to start session");
        }

        sessionData = await res.json();
        setActiveSession(sessionData.ID);

        if (sessionData.Steps && sessionData.Steps.length > 0) {
            displaySteps(sessionData.Steps);
        }

        connectToSession();
        const targetTime = formatTime(sessionData.TargetSec);
        document.getElementById("activeStep").textContent = `Session started - target: ${targetTime}`;
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
        if (sessionData && sessionData.Steps && sessionData.Steps[data.index]) {
            const step = sessionData.Steps[data.index];


            if (step.Duration - data.elapsed === 10) {
                notifyUser(
                    "Almost done!",
                    `10 seconds left on step ${data.index + 1}`
                );
            }
            if (data.elapsed >= step.Duration) {
                notifyUser(
                    "Step complete!",
                    `Step ${data.index + 1} finished!`
                );
            }
        }
    };
    es.onerror = (err) => {
        console.log("EventSource connection closed:", err);
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
        document.getElementById("activeStep").textContent = "Session stopped, (not saved)";
        document.getElementById("steps").innerHTML = "";
        if (es) es.close();
        setActiveSession(null);
        sessionData = null;
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
        document.getElementById("steps").innerHTML = "";
        if (es) es.close();

        setActiveSession(null);
        sessionData = null;

        hydrateTargetInput();

    } catch (err) {
        console.error(err);
        alert(err.message);
    }
}
