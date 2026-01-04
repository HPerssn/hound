let es;
let sessionId;

async function startSession() {
    const targetSec = parseInt(document.getElementById("targetSec").value, 10);
    if (!targetSec) return;

    try {
        const res = await fetch('/sessions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ targetSec }) // only targetSec, ID is generated server-side
        });

        if (!res.ok) throw new Error("failed to start session");

        const session = await res.json();
        sessionId = session.ID;

        connectToSession();
    } catch (err) {
        console.error(err);
    }
}

function connectToSession() {
    if (!sessionId) return;
    if (es) es.close();

    es = new EventSource(`/sessions/${sessionId}/events`);

    es.onmessage = (e) => {
        const data = JSON.parse(e.data);

        // active step timer
        if (data.index !== undefined && data.elapsed !== undefined) {
            document.getElementById("activeStep").textContent =
                `Step ${data.index + 1}: ${data.elapsed}s elapsed`;
        }

        // log step completion from Step.Completed
        if (data.completed) {
            document.getElementById("activeStep").textContent =
                `Step ${data.index + 1} completed`;
            document.getElementById("log").textContent +=
                JSON.stringify(data, null, 2) + "\n\n";
        }

        // session done
        if (data.type === "done") {
            document.getElementById("activeStep").textContent = "Session finished";
        }
    };

    es.onerror = () => es.close();
}

async function stopSession() {
    if (!sessionId) return;
    await fetch(`/sessions/${sessionId}/stop`, { method: 'POST' });
    document.getElementById("activeStep").textContent = "Session stopped";
}
