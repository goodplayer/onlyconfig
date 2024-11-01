import {OnlyConfigNavBar} from "../../components/header.jsx";
import {IsLogin, LoginToken} from "../../components/loginstatus.jsx";
import Redirect from "../../components/redirect.jsx";
import {useEffect, useState} from "preact/hooks";
import {ApiEndpoint} from "../../components/onlyconfig_configure.jsx";

export function OrgManagement() {
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

    let [ownedOrgList, setOwnedOrgList] = useState(null);
    let [selectedOrg, setSelectedOrg] = useState(null);
    let prepareRefresh = function () {
        setOwnedOrgList(null);
        setSelectedOrg(null);
    }

    // load org data
    let loadOrgDataFn = function () {
        // clear old data first
        prepareRefresh();
        fetch(ApiEndpoint('/user/organizations'), {
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
                    alert("Get organizations data error");
                    return
                }
                let dataJson = await res.json();
                console.log("fetch organizations data:", dataJson.result)
                if (!dataJson.result) {
                    setOwnedOrgList([]);
                } else {
                    setOwnedOrgList(dataJson.result);
                }
            })
            .catch(e => {
                console.log("fetch organizations data failed:", e)
                //FIXME better error information display
                alert("Get organizations data error");
            })
    }
    // submit new org
    let submitNewOrgForm = function (event) {
        let org = encodeURI(event.target.org.value);
        event.target.org.value = '';
        let url = `/organization/${org}`
        fetch(ApiEndpoint(url), {
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
                    alert("put new organization error");
                    return
                }
                // trigger organization fetching
                loadOrgDataFn();
            })
            .catch(e => {
                console.log("put new organization failed:", e)
                //FIXME better error information display
                alert("put new organization error");
            })
    }
    // submit add owner to organization
    let submitAddOwnerForm = function (event) {
        let org_name = selectedOrg.org_name;
        let username = event.target.username.value;
        let url = `/organization/${org_name}/owner/${username}`;
        fetch(ApiEndpoint(url), {
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
                    alert("put add user to org error");
                    return
                }
                // trigger organization fetching
                loadOrgDataFn();
            })
            .catch(e => {
                console.log("put add user to org failed:", e)
                //FIXME better error information display
                alert("put add user to org error");
            })
    }

    useEffect(() => {
        loadOrgDataFn();
    }, []);

    return (
        <>
            <OnlyConfigNavBar/>
            <div className={"row"} style={"margin:5px 0px;"}>
                <div className={"col-1"}></div>
                <div className={"col-10 border rounded"}>
                    <h1>Manage owned organization</h1>
                    <div className={"row border rounded"} style={"margin: 10px 0px; padding: 5px 5px;"}>
                        <h4>Create new organization</h4>
                        <form onSubmit={event => {
                            event.preventDefault();
                            submitNewOrgForm(event)
                        }}>
                            <h5>Organization name:</h5>
                            <input type="text" className="form-control" name={"org"} autocomplete={"off"}/>
                            <hr/>
                            <button type="submit" className="btn btn-primary">Add</button>
                        </form>
                    </div>
                    <div className={"row border rounded"} style={"margin: 10px 0px; padding: 5px 5px;"}>
                        <h4>Manage organization</h4>
                        {!ownedOrgList && <>Loading...</>}
                        {ownedOrgList && <>
                            <select className="form-select" onChange={(event) => {
                                setSelectedOrg(ownedOrgList[event.target.value]);
                            }}>
                                <option disabled selected value> -- select an organization --</option>
                                {ownedOrgList.map((elem, index) => (
                                    <option value={index}>{elem.org_name}</option>
                                ))}
                            </select>
                        </>}
                        {selectedOrg && <>
                            <div className={"row"} style={"margin: 5px 0px;"}>
                                <div className={"col-6"}>
                                    <h5>Owner list</h5>
                                    <ul className="list-group">
                                        {selectedOrg.owner_list && selectedOrg.owner_list.map(elem => (
                                            <li className="list-group-item">{elem}</li>
                                        ))}
                                    </ul>
                                </div>
                                <div className={"col-6"}>
                                    <h5>User list</h5>
                                    <ul className="list-group">
                                        {selectedOrg.user_list && selectedOrg.user_list.map(elem => (
                                            <li className="list-group-item">{elem}</li>
                                        ))}
                                    </ul>
                                </div>
                            </div>
                            <div className={"row"} style={"margin: 5px 0px;"}>
                                <div className={"col-6"}>
                                    <hr/>
                                    <span>Add owner:</span>
                                    <form onSubmit={(event) => {
                                        event.preventDefault();
                                        submitAddOwnerForm(event);
                                    }}>
                                        <input type="text" className="form-control" name="username"
                                               autocomplete={"off"}/>
                                        <button type="submit" className="btn btn-primary" style={"margin: 5px 0px;"}>Add
                                        </button>
                                    </form>
                                </div>
                                <div className={"col-6"}>
                                    <hr/>
                                    <span>Add user:</span>
                                    <input type="text" className="form-control" name="username" autocomplete={"off"}
                                           disabled={true}/>
                                    <button type="button" className="btn btn-primary" style={"margin: 5px 0px;"}
                                            disabled={true}>Add
                                    </button>
                                </div>
                            </div>
                        </>}
                    </div>
                </div>
                <div className={"col-1"}></div>
            </div>
        </>
    )
}
