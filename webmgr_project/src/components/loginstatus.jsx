const userSessionStorageKey = 'onlyconfig.user';

export function IsLogin() {
    let userString = localStorage.getItem(userSessionStorageKey)
    if (userString === null) {
        return false;
    }
    let user = JSON.parse(userString);
    return user.is_login === true;
}

export function SetLoggedIn(token) {
    let user = {
        'is_login': true,
        'token': token,
    }
    localStorage.setItem(userSessionStorageKey, JSON.stringify(user));
}

export function SetLoggedOut() {
    localStorage.removeItem(userSessionStorageKey);
}

export function LoginToken() {
    let userString = localStorage.getItem(userSessionStorageKey)
    if (userString === null) {
        return "";
    }
    let user = JSON.parse(userString);
    return user.token;
}
