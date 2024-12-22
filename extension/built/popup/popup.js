document.getElementById('captureBtn').addEventListener('click', function () {
    browser.tabs.query({ active: true, currentWindow: true })
        .then(function (tabs) {
        if (tabs[0]) {
            return browser.tabs.executeScript(tabs[0].id, { file: "content.js" })
                .then(function () {
                return new Promise(function (resolve) { return setTimeout(resolve, 100); });
            })
                .then(function () {
                return browser.tabs.sendMessage(tabs[0].id, { action: "startCapture" });
            })
                .catch(function (error) {
                console.error("Error in capture process:", error);
            });
        }
        else {
            console.error("No active tab found");
        }
    })
        .catch(function (error) {
        console.error("Error in capture process:", error);
    });
});
