export function ApiEndpoint(url) {
    if (!url.startsWith('/')) {
        url = '/' + url;
    }
    //FIXME need to configure server endpoint
    return 'http://localhost:8880' + url
}
