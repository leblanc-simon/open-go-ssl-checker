function formatISODateToReadable(isoDateStr) {
    if (!isoDateStr) return '-';
    try {
        const date = new Date(isoDateStr);
        const day = String(date.getDate()).padStart(2, '0');
        const month = date.toLocaleString('default', { month: 'short' });
        const year = date.getFullYear();
        const hours = String(date.getHours()).padStart(2, '0');
        const minutes = String(date.getMinutes()).padStart(2, '0');
        return `${day} ${month} ${year} ${hours}:${minutes}`;
    } catch (e) {
        console.error("Error formatting date:", isoDateStr, e);
        return isoDateStr;
    }
}

function updateTable(summaries) {
    const tbody = document.querySelector("table tbody");
    if (!tbody) {
        console.error("Table body not found!");
        return;
    }
    tbody.innerHTML = "";

    if (!summaries || summaries.length === 0) {
        const container = document.querySelector("main");
        const noDataPara = document.createElement('p');
        noDataPara.className = 'no-data';
        noDataPara.innerHTML = 'Aucun projet surveillé pour le moment. <a href="/add">Ajoutez-en un !</a>';

        const existingNoData = container.querySelector('.no-data');
        if(existingNoData) existingNoData.remove();

        const table = container.querySelector('table');
        if(table) table.style.display = 'none';

        container.appendChild(noDataPara);
        return;
    } else {
            const container = document.querySelector("main");
            const existingNoData = container.querySelector('.no-data');
            if(existingNoData) existingNoData.remove();
            const table = container.querySelector('table');
            if(table) table.style.display = '';
    }


    summaries.forEach(summary => {
        const row = tbody.insertRow();

        row.insertCell().textContent = summary.ProjectName || '-';
        row.insertCell().textContent = (summary.Host && summary.Port) ? `${summary.Host}:${summary.Port}` : '-';
        row.insertCell().textContent = summary.Type ? summary.Type.toUpperCase() : '-';

        const checkTimeCell = row.insertCell();
        checkTimeCell.textContent = summary.CheckTime ? formatISODateToReadable(summary.CheckTime) : '-';
        if (!summary.CheckTime) checkTimeCell.classList.add('no-data');


        const domainsCell = row.insertCell();
        domainsCell.textContent = summary.Domains || '-';
        if (!summary.Domains) domainsCell.classList.add('no-data');

        const ipCell = row.insertCell();
        ipCell.textContent = summary.IP || '-';
        if (!summary.IP) ipCell.classList.add('no-data');

        const issuerCell = row.insertCell();
        issuerCell.textContent = summary.Issuer || '-';
        if (!summary.Issuer) issuerCell.classList.add('no-data');

        const expiryDateCell = row.insertCell();
        expiryDateCell.textContent = summary.ExpiryDate || '-';
        if (!summary.ExpiryDate) expiryDateCell.classList.add('no-data');

        const daysRemainingCell = row.insertCell();
        if (summary.DaysRemaining !== null && summary.DaysRemaining !== undefined) {
            const days = summary.DaysRemaining;
            const span = document.createElement('span');
            span.textContent = days;
            if (days < 0) { span.className = "days-critical"; span.title="Expiré"; }
            else if (days < 15) { span.className = "days-critical"; }
            else if (days < 30) { span.className = "days-warning"; }
            else { span.className = "days-ok"; }
            daysRemainingCell.appendChild(span);
        } else {
            daysRemainingCell.textContent = '-';
            daysRemainingCell.classList.add('no-data');
        }

        const actionsCell = row.insertCell();
        const link = document.createElement('a');
        link.href = `/history/${summary.ProjectID}`;
        link.className = "action-link link-details";
        link.textContent = "Historique";
        actionsCell.appendChild(link);
    });
}

function connectWebSocket()
{
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const socketURL = `${protocol}//${window.location.host}/ws`;
    const socket = new WebSocket(socketURL);

    socket.onopen = () => {
        console.log("WebSocket connection established.");
    };

    socket.onmessage = (event) => {
        try {
            console.log("WebSocket message received:", event.data);
            const summaries = JSON.parse(event.data);
            updateTable(summaries);
        } catch (e) {
            console.error("Error parsing WebSocket message or updating table:", e);
        }
    };

    socket.onclose = (event) => {
        console.log("WebSocket connection closed.", event);
        // Vous pourriez vouloir tenter une reconnexion ici
        setTimeout(connectWebSocket, 5000); // Exemple de reconnexion après 5s
    };

    socket.onerror = (error) => {
        console.error("WebSocket error:", error);
    };

    return socket;
}

document.addEventListener("DOMContentLoaded", () => {
    const socket = connectWebSocket();

    const refreshButton = document.getElementById("refreshButton");

    if (refreshButton) {
        refreshButton.onclick = () => {
            if (socket && socket.readyState === WebSocket.OPEN) {
                console.log("Sending 'refresh' message to WebSocket server.");
                socket.send("refresh");
            } else {
                console.log("WebSocket is not connected. Cannot send 'refresh'.");
                // Peut-être essayer de se reconnecter ou informer l'utilisateur
            }
        };
    }
});
