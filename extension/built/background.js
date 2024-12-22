browser.runtime.onMessage.addListener(function (request, sender, sendResponse) {
    if (request.action === "captureRegion") {
        captureRegion(request.region);
    }
});
function captureRegion(region) {
    browser.tabs.captureVisibleTab(null, { format: "png" }).then(function (dataUrl) {
        var image = new Image();
        image.onload = function () {
            var canvas = document.createElement('canvas');
            var ctx = canvas.getContext('2d');
            canvas.width = region.width;
            canvas.height = region.height;
            ctx.drawImage(image, region.x, region.y, region.width, region.height, 0, 0, region.width, region.height);
            var croppedDataUrl = canvas.toDataURL('image/png');
            sendToServer(croppedDataUrl);
        };
        image.src = dataUrl;
    });
}
function sendToServer(dataUrl) {
    var binaryString = atob(dataUrl.split(',')[1]);
    var len = binaryString.length;
    var bytes = new Uint8Array(len);
    for (var i = 0; i < len; i++) {
        bytes[i] = binaryString.charCodeAt(i);
    }
    // Send to Go server
    browser.runtime.sendNativeMessage("com.example.goscreenshot", { action: "saveScreenshot", data: Array.from(bytes) }, function (response) {
        console.log("Response from Go:", response);
    });
}
