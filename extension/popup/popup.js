document.getElementById('captureBtn').addEventListener('click', () => {
    browser.tabs.query({ active: true, currentWindow: true })
        .then(tabs => {
            if (tabs[0]) {
                return browser.tabs.executeScript(tabs[0].id, { file: "content.js" })
                    .then(() => {
                        return new Promise(resolve => setTimeout(resolve, 100));
                    })
                    .then(() => {
                        return browser.tabs.sendMessage(tabs[0].id, { action: "startCapture" });
                    })
                    .catch(error => {
                        console.error("Error in capture process:", error);
                    });
            } else {
                console.error("No active tab found");
            }
        })
        .catch(error => {
            console.error("Error in capture process:", error);
        });
});
