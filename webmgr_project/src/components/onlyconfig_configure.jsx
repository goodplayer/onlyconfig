export function ApiEndpoint(url) {
    if (!url.startsWith('/')) {
        url = '/' + url;
    }
    // development url
    //FIXME use better way to support development server url
    try {
        if (process.env.NODE_ENV === "development") {
            return 'http://localhost:8880' + url
        }
    } catch (e) {
        console.log("determine env failed.")
    }
    return "//" + window.location.host + url
}
