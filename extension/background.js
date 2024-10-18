browser.runtime.onMessage.addListener((request, sender, sendResponse) => {
    if (request.action === "captureRegion") {
        captureRegion(request.region);
    }
})

function captureRegion(region) {
    browser.tabs.captureVisibleTab(null, { format: "png" }).then((dataUrl) => {
        const image = new Image();
        image.onload = () => {
            const canvas = document.createElement('canvas');
            const ctx = canvas.getContext('2d');
            canvas.width = region.width;
            canvas.height = region.height;
            ctx.drawImage(image, region.x, region.y, region.width, region.height, 0, 0, region.width, region.height);

            const croppedDataUrl = canvas.toDataURL('image/png');
            sendToServer(croppedDataUrl);
        };
        image.src = dataUrl;
    });
}


function sendToServer(dataUrl) {
    const binaryString = atob(dataUrl.split(',')[1]);
    const len = binaryString.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) {
        bytes[i] = binaryString.charCodeAt(i);
    }

    // Send to Go server
    browser.runtime.sendNativeMessage(
        "com.example.goscreenshot",
        { action: "saveScreenshot", data: Array.from(bytes) },
        (response) => {
            console.log("Response from Go:", response);
        }
    );
}
