import {useState} from "preact/hooks";
import {IsLogin, LoginToken, SetLoggedIn, SetLoggedOut} from '../../components/loginstatus.jsx';
import Redirect from "../../components/redirect.jsx";
import {ApiEndpoint} from "../../components/onlyconfig_configure.jsx";
import {OnlyConfigNavBar} from "../../components/header.jsx";

export function Logout() {
    SetLoggedOut();

    return (
        <Redirect to='/'/>
    );
}

export function LoginPage() {
    let [formData, setFormData] = useState({});
    let [errorDisplay, setErrorDisplay] = useState('');

    let [loggedIn, setLoggedIn] = useState(false);

    let onSubmit = e => {
        console.log("submit!!!!")
        // components states reset
        e.submitter.disabled = true
        e.preventDefault();
        setErrorDisplay('');

        fetch(ApiEndpoint('/auth/user/login'), {
            headers: {
                'content-type': 'application/json; charset=UTF-8'
            },
            method: 'post',
            body: JSON.stringify({
                username: formData.username,
                password: formData.password,
            }),
        })
            .then(async res => {
                if (res.status === 401) {
                    setErrorDisplay('Invalid username or password')
                    e.submitter.disabled = false;
                    return
                }
                if (res.status !== 200) {
                    console.log("status:", res.status);
                    e.submitter.disabled = false;
                    return
                }
                // trigger main page redirect by setting up login status
                let tk = await res.json();
                SetLoggedIn(tk.token);
                setLoggedIn(true);
            })
            .catch(async e => {
                console.log("send request err:", e)
                e.submitter.disabled = false
            })
    }

    let onFormChange = e => {
        //FIXME not working since just editing properties of the same object when using useState
        var newData = formData;
        newData[e.target.name] = e.target.value;
        setFormData(newData);
        // console.log("new data:", formData);
    }

    if (loggedIn) {
        return (
            <Redirect to='/'/>
        );
    }

    let [hitRegister, setHitRegister] = useState(null);
    if (hitRegister) {
        return (
            <Redirect to={"/register"}/>
        )
    }

    return (
        <div className="position-absolute top-50 start-50 translate-middle shadow border rounded"
             style={"padding: 10px 10px"}>
            <form onSubmit={onSubmit}>
                <div><h4>OnlyConfig Web Manager Login</h4></div>
                <div>
                    <label htmlFor="login_username" className="form-label">Username</label>
                    <input className="form-control" type="text" id="login_username" value={formData.username}
                           name="username" autoComplete="off" onChange={onFormChange}/>
                </div>
                <div>
                    <label htmlFor="login_password" className="form-label">Password</label>
                    <input className="form-control" type="password" id="login_password" value={formData.password}
                           name="password" autoComplete="off" onChange={onFormChange}/>
                </div>
                {
                    errorDisplay !== '' && <div>
                        <div><label style="color: red;">{errorDisplay}</label></div>
                    </div>
                }
                <div style={"margin-top: 10px;"}>
                    <button type="submit" className="btn btn-primary" style={"width: 100px"}>Login
                    </button>
                    <span style={"padding: 0 10px"}><a href={"#"} onClick={() => {
                        setHitRegister(true);
                    }}>Register New User
                    </a></span>
                </div>
            </form>
        </div>
    );
}

export function RegisterPage() {

    let [errorDisplay, setErrorDisplay] = useState(null);

    let [hitLogin, setHitLogin] = useState(null);
    if (hitLogin) {
        return (
            <Redirect to={'/login'}/>
        )
    }

    let onSubmit = function (event) {
        event.preventDefault();
        setErrorDisplay(null);
        event.submitter.disabled = true

        let username = event.target.username.value;
        let password = event.target.password.value;
        let confirmedPassword = event.target.confirm_password.value;
        let email = event.target.email.value;
        let displayName = event.target.display_name.value;

        if (password !== confirmedPassword) {
            setErrorDisplay("password mismatch");
            event.submitter.disabled = false;
            return
        }

        fetch(ApiEndpoint("/user/new_user"), {
            headers: {
                'content-type': 'application/json; charset=UTF-8'
            },
            method: 'post',
            body: JSON.stringify({
                username: username,
                password: password,
                email: email,
                display_name: displayName,
            }),
        })
            .then(async res => {
                if (res.status !== 200) {
                    console.log("status:", res.status);
                    setErrorDisplay('register failed');
                    event.submitter.disabled = false;
                    return
                }
                // jump to login page
                setHitLogin(true);
            })
            .catch(async e => {
                console.log("send request err:", e)
                event.submitter.disabled = false
            })
    }

    return (
        <div className="position-absolute top-50 start-50 translate-middle shadow border rounded"
             style={"padding: 10px 10px"}>
            <form onSubmit={onSubmit}>
                <div><h4>OnlyConfig Web Manager Register</h4></div>
                <div>
                    <label htmlFor="reg_username" className="form-label">Username</label>
                    <input className="form-control" type="text" id="reg_username" name="username" autoComplete="off"/>
                </div>
                <div>
                    <label htmlFor="reg_password" className="form-label">Password</label>
                    <input className="form-control" type="password" id="reg_password" name="password"
                           autoComplete="off"/>
                </div>
                <div>
                    <label htmlFor="reg_confirmed_password" className="form-label">Confirm Password</label>
                    <input className="form-control" type="password" id="reg_confirmed_password"
                           name="confirm_password" autoComplete="off"/>
                </div>
                <div>
                    <label htmlFor="reg_email" className="form-label">Email</label>
                    <input className="form-control" type="email" id="reg_email" name="email" autoComplete="off"/>
                </div>
                <div>
                    <label htmlFor="reg_display_name" className="form-label">Display Name</label>
                    <input className="form-control" type="text" id="reg_display_name" name="display_name"
                           autoComplete="off"/>
                </div>
                {
                    errorDisplay !== '' && <div>
                        <div><label style="color: red;">{errorDisplay}</label></div>
                    </div>
                }
                <div style={"margin-top: 10px;"}>
                    <button type="submit" className="btn btn-primary" style={"width: 100px"}>Register
                    </button>
                    <span style={"padding: 0 10px"}>
                        <a href={"#"} onClick={() => {
                            setHitLogin(true);
                        }}>Back to login
                        </a>
                    </span>
                </div>
            </form>
        </div>
    );
}

export function ChangePassword() {
    if (!IsLogin()) {
        return (
            <Redirect to='/login'/>
        );
    }
    let [isLogout, setLogout] = useState(false);
    if (isLogout) {
        return (
            <Redirect to='/logout'/>
        );
    }

    let [errorDisplay, setErrorDisplay] = useState(null);

    let onSubmit = function (event) {
        event.preventDefault();
        setErrorDisplay(null);

        let oldpwd = event.target.old.value;
        let newpwd = event.target.new.value;
        let confirm = event.target.confirm.value;
        if (newpwd !== confirm) {
            setErrorDisplay("password mismatch");
        }
        event.submitter.disabled = true

        fetch(ApiEndpoint("/user/change_password"), {
            headers: {
                'content-type': 'application/json; charset=UTF-8',
                'Authorization': 'Bearer ' + LoginToken(),
            },
            method: 'post',
            body: JSON.stringify({
                old: oldpwd,
                new: newpwd,
            }),
        })
            .then(async res => {
                if (res.status !== 200) {
                    console.log("status:", res.status);
                    setErrorDisplay('change password failed');
                    event.submitter.disabled = false;
                    return
                }
                // clear form values
                event.target.old.value = '';
                event.target.new.value = '';
                event.target.confirm.value = '';
            })
            .catch(async e => {
                console.log("send request err:", e)
                event.submitter.disabled = false
            })
    }

    return (
        <>
            <OnlyConfigNavBar/>
            <div className="row" style="padding: 10px 0">
                <div className="col-1"></div>
                <div className="col-10 border rounded">
                    <h1>Change password</h1>
                    <div className="row">
                        <div className="col-1"></div>
                        <div className="col-10">
                            <form onSubmit={onSubmit}>
                                <div>Old password</div>
                                <div className="btn-group" style="padding: 10px 0">
                                    <input className="form-control form-control-lg" type="password" name="old"
                                           autoComplete='off'/>
                                </div>
                                <div>New password</div>
                                <div className="btn-group" style="padding: 10px 0">
                                    <input className="form-control form-control-lg" type="password" name="new"
                                           autoComplete='off'/>
                                </div>
                                <div>Confirm password</div>
                                <div className="btn-group" style="padding: 10px 0">
                                    <input className="form-control form-control-lg" type="password" name="confirm"
                                           autoComplete='off'/>
                                </div>
                                {
                                    errorDisplay && <div>
                                        <div style="color: red;">{errorDisplay}</div>
                                    </div>
                                }
                                <div style={"margin: 10px 0px"}>
                                    <button type="submit" className="btn btn-primary">Change</button>
                                </div>
                            </form>
                        </div>
                        <div className="col-1"></div>
                    </div>
                </div>
                <div className="col-1"></div>
            </div>
        </>
    );
}
