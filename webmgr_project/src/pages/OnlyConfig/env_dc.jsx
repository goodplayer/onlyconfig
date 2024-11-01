import {OnlyConfigNavBar} from "../../components/header.jsx";
import {IsLogin, LoginToken} from "../../components/loginstatus.jsx";
import Redirect from "../../components/redirect.jsx";
import {useEffect, useState} from "preact/hooks";
import {ApiEndpoint} from "../../components/onlyconfig_configure.jsx";

export function EnvAndDc() {
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

    // fetch env and dc when loading
    let [dataFetched, setDataFetched] = useState(null);
    let loadDataFn = function () {
        fetch(ApiEndpoint('/configures/env_dc_list'), {
            headers: {
                'content-type': 'application/json; charset=UTF-8',
                'Authorization': 'Bearer ' + LoginToken(),
            },
            method: 'get',
        })
            .then(async res => {
                if (res.status === 401) {
                    setLogout(true);
                    return
                }
                if (res.status !== 200) {
                    console.log("status:", res.status)
                    //FIXME better error information display
                    alert("Get env_dc_list error");
                    return
                }
                let dataJson = await res.json();
                setDataFetched(dataJson);
            })
            .catch(e => {
                console.log("fetch env_dc_list failed:", e)
                //FIXME better error information display
                alert("Get env_dc_list error");
            })
    }
    useEffect(() => {
        //load env and datacenter at first
        loadDataFn();
        return () => {
        };
    }, []);
    //FIXME add placeholder when data is loading
    if (!dataFetched) {
        return (
            <>
                <OnlyConfigNavBar/>
                <div>loading....</div>
            </>
        );
    }

    // add new env or dc submission function
    let handleSubmitNew = function (addType, addName) {
        fetch(ApiEndpoint('/configures/env_and_dc/' + encodeURI(addType) + '/' + encodeURI(addName)), {
            headers: {
                'content-type': 'application/json; charset=UTF-8',
                'Authorization': 'Bearer ' + LoginToken(),
            },
            method: 'put',
        })
            .then(async res => {
                if (res.status === 401) {
                    setLogout(true);
                    return
                }
                if (res.status !== 200) {
                    console.log("status:", res.status)
                    //FIXME better error information display
                    alert("put env or dc error");
                    return
                }
                // trigger new data fetching
                loadDataFn();
            })
            .catch(e => {
                console.log("put env or dc failed:", e)
                //FIXME better error information display
                alert("put env or dc error");
            })
    }

    return (
        <>
            <OnlyConfigNavBar/>
            <div className="row" style="padding: 10px 0">
                <div className="col-1"></div>
                <div className="col-10 border rounded">
                    <h1>Add environment</h1>
                    <div className="row">
                        <div className="col-1"></div>
                        <div className="col-10">
                            <strong>Current environment:</strong>
                            <ul>
                                {dataFetched.env && dataFetched.env.map(elem => (
                                        <li>{elem}</li>
                                    )
                                )}
                            </ul>
                            <form onSubmit={(e) => {
                                e.preventDefault();
                                handleSubmitNew('env', e.target.env.value);
                                e.target.env.value = '';
                            }}>
                                <div className="btn-group" style="padding: 10px 0">
                                    <input className="form-control form-control-lg" type="text" name="env"
                                           autocomplete='off'/>
                                    <button type="submit" className="btn btn-primary">Add</button>
                                </div>
                            </form>
                        </div>
                        <div className="col-1"></div>
                    </div>
                </div>
                <div className="col-1"></div>
            </div>
            <div className="row" style="padding: 10px 0">
                <div className="col-1"></div>
                <div className="col-10 border rounded">
                    <h1>Add datacenter</h1>
                    <div className="row">
                        <div className="col-1"></div>
                        <div className="col-10">
                            <strong>Current datacenter:</strong>
                            <ul>
                                {dataFetched.dc && dataFetched.dc.map(elem => (
                                        <li>{elem}</li>
                                    )
                                )}
                            </ul>
                            <form onSubmit={(e) => {
                                e.preventDefault();
                                handleSubmitNew('dc', e.target.dc.value);
                                e.target.dc.value = '';
                            }}>
                                <div className="btn-group" style="padding: 10px 0">
                                    <input className="form-control form-control-lg" type="text" name="dc"
                                           autocomplete='off'/>
                                    <button type="submit" className="btn btn-primary">Add</button>
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