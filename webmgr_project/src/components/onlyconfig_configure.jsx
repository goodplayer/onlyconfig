export function ApiEndpoint(url) {
    if (!url.startsWith('/')) {
        url = '/' + url;
    }
    //FIXME need to configure server endpoint
    // console.log("current domain:", "//" + window.location.host);
    // return 'http://localhost:8880' + url
    return "//" + window.location.host + url
}
