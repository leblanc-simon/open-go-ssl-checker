const updateDefaultPort = () => {
    const protocolSelect = document.getElementById('type');
    const portInput = document.getElementById('port');
    const selectedOption = protocolSelect.options[protocolSelect.selectedIndex];
    const defaultPort = selectedOption.getAttribute('data-default');
    if (defaultPort) {
        portInput.value = defaultPort;
    }
}

document.addEventListener('DOMContentLoaded', function() {
    const protocolSelect = document.getElementById('type');
    updateDefaultPort();
    protocolSelect.addEventListener('change', updateDefaultPort);
});