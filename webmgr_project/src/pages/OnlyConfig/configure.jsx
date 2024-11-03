import {OnlyConfigNavBar} from "../../components/header.jsx";
import {ApiEndpoint} from "../../components/onlyconfig_configure.jsx";
import {IsLogin, LoginToken} from "../../components/loginstatus.jsx";
import {useEffect, useState} from "preact/hooks";
import Redirect from "../../components/redirect.jsx";

export function OnlyConfigHome() {

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

    // selected items
    let [selectedApp, setSelectedApp] = useState(null); // [app_id, app_name]
    let [selectedEnvAndDc, setSelectedEnvAndDc] = useState(null);
    let clearSelected = function () {
        setSelectedApp(null);
        setSelectedEnvAndDc(null);
    }

    // load application data
    let [applications, setApplications] = useState(null);
    let loadApplicationData = function (callback) {
        fetch(ApiEndpoint('/configures/applications'), {
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
                    alert("Get application data error");
                    return
                }
                let dataJson = await res.json();
                console.log("fetch application data:", dataJson.list)
                if (!dataJson.list) {
                    setApplications([]);
                } else {
                    setApplications(dataJson.list);
                }
                if (callback) {
                    callback();
                }
            })
            .catch(e => {
                console.log("fetch application data failed:", e)
                //FIXME better error information display
                alert("Get application data error");
            })
    }
    useEffect(() => {
        loadApplicationData();
        return () => {
        };
    }, []);
    //FIXME add placeholder when data is loading
    if (!applications) {
        return (
            <>
                <OnlyConfigNavBar/>
                <div>loading....</div>
            </>
        );
    }

    // on display createNewAppModal, load org data
    let [orgList, setOrgList] = useState(null);
    let [envList, setEnvList] = useState(null);
    useEffect(() => {
        let createNewAppModalFn = function (event) {
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
                        setOrgList([]);
                    } else {
                        setOrgList(dataJson.result);
                    }
                })
                .catch(e => {
                    console.log("fetch organizations data failed:", e)
                    //FIXME better error information display
                    alert("Get organizations data error");
                })
        }
        let addDatacenterModelFn = function (event) {
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
                        alert("Get environments and datacenters data error");
                        return
                    }
                    let dataJson = await res.json();
                    console.log("fetch environments and datacenters data:", dataJson)
                    if (!dataJson) {
                        setEnvList({});
                    } else {
                        setEnvList(dataJson);
                    }
                })
                .catch(e => {
                    console.log("fetch environments and datacenters data failed:", e)
                    //FIXME better error information display
                    alert("Get environments and datacenters data error");
                })
        }
        document.getElementById('createNewAppModal').addEventListener('show.bs.modal', createNewAppModalFn);
        document.getElementById('addDatacenterModal').addEventListener('show.bs.modal', addDatacenterModelFn);
        return () => {
            document.getElementById('createNewAppModal').removeEventListener('show.bs.modal', createNewAppModalFn);
            document.getElementById('addDatacenterModal').removeEventListener('show.bs.modal', addDatacenterModelFn);
        }
    }, []);
    // submit creating new app form
    let submitCreatingNewAppForm = function (event) {
        // handle creating new app
        let orgId = event.target.org.value;
        let appName = event.target.app.value;
        event.target.app.value = '';
        fetch(ApiEndpoint('/configures/application/' + encodeURI(orgId) + '/' + encodeURI(appName)), {
            headers: {
                'content-type': 'application/json; charset=UTF-8',
                'Authorization': 'Bearer ' + LoginToken(),
            },
            method: 'put',
        })
            .then(async res => {
                console.log("created status:", res.status)
                if (res.status === 401) {
                    setLogout(true);
                    return
                }
                if (res.status !== 200) {
                    console.log("status:", res.status)
                    //FIXME better error information display
                    alert("put new application error");
                    return
                }
                // trigger new data fetching and close the modal when complete
                loadApplicationData(function () {
                    bootstrap.Modal.getInstance(document.getElementById('createNewAppModal')).hide();
                });
            })
            .catch(e => {
                console.log("put new application failed:", e)
                //FIXME better error information display
                alert("put new application error");
            })
    }
    // submit add datacenter form
    let submitAddDatacenterForm = function (event) {
        console.log("submit:", event.target.env.value)
        console.log("submit:", event.target.dc.value)
        console.log("selected app:", selectedApp)
        let app_id = encodeURI(selectedApp[0]);
        let env = encodeURI(event.target.env.value);
        let dc = encodeURI(event.target.dc.value);
        fetch(ApiEndpoint('/configures/application/' + app_id + '/' + env + '/' + dc), {
            headers: {
                'content-type': 'application/json; charset=UTF-8',
                'Authorization': 'Bearer ' + LoginToken(),
            },
            method: 'put',
        })
            .then(async res => {
                console.log("created status:", res.status)
                if (res.status === 401) {
                    setLogout(true);
                    return
                }
                if (res.status !== 200) {
                    console.log("status:", res.status)
                    //FIXME better error information display
                    alert("put datacenter for app error");
                    return
                }
                // trigger new data fetching and close the modal when complete
                loadApplicationData(function () {
                    bootstrap.Modal.getInstance(document.getElementById('addDatacenterModal')).hide();
                });
            })
            .catch(e => {
                console.log("put datacenter for app failed:", e)
                //FIXME better error information display
                alert("put datacenter for app error");
            })
    }

    //FIXME Question1: Does 'height:100vh' setting allow elements exceeding the max height?
    return (
        <div>
            <OnlyConfigNavBar/>

            <div className="modal fade" id="createNewAppModal" data-bs-backdrop="static" data-bs-keyboard="false"
                 tabIndex="-1" aria-labelledby="createNewAppModalLabel" aria-hidden="true">
                <div className="modal-dialog modal-lg">
                    <div className="modal-content">
                        <div className="modal-header">
                            <h1 className="modal-title fs-5" id="createNewAppModalLabel">New Application</h1>
                            <button type="button" className="btn-close" data-bs-dismiss="modal"
                                    aria-label="Close"></button>
                        </div>
                        <div className="modal-body">
                            <form onSubmit={event => {
                                event.preventDefault();
                                submitCreatingNewAppForm(event);
                            }}>
                                {orgList && <>
                                    <h5>Organization</h5>
                                    <select className="form-select" name="org">
                                        {orgList.map(elem => (
                                            <option value={elem.org_id}>{elem.org_name}</option>
                                        ))}
                                    </select>
                                    <h5>Application name</h5>
                                    <input type="text" name="app" className="form-control" autoComplete="off"/>
                                    <hr/>
                                    <button type="submit" className="btn btn-primary">Add</button>
                                </>}
                                {!orgList && <div>loading...</div>}
                            </form>
                        </div>
                        <div className="modal-footer">
                            <button type="button" className="btn btn-secondary" data-bs-dismiss="modal">Close</button>
                        </div>
                    </div>
                </div>
            </div>

            <div className="modal fade" id="addDatacenterModal" data-bs-backdrop="static" data-bs-keyboard="false"
                 tabIndex="-1" aria-labelledby="addDatacenterModalLabel" aria-hidden="true">
                <div className="modal-dialog modal-lg">
                    <div className="modal-content">
                        <div className="modal-header">
                            <h1 className="modal-title fs-5" id="addDatacenterModalLabel">Add datacenter</h1>
                            <button type="button" className="btn-close" data-bs-dismiss="modal"
                                    aria-label="Close"></button>
                        </div>
                        <div className="modal-body">
                            <form onSubmit={event => {
                                event.preventDefault();
                                submitAddDatacenterForm(event);
                            }}>
                                {envList && envList.env && envList.dc && <>
                                    <h5>Environment</h5>
                                    <select className="form-select" name="env">
                                        {envList.env.map(elem => (
                                            <option value={elem}>{elem}</option>
                                        ))}
                                    </select>
                                    <h5>Datacenter</h5>
                                    <select className="form-select" name="dc">
                                        {envList.dc.map(elem => (
                                            <option value={elem}>{elem}</option>
                                        ))}
                                    </select>
                                    <hr/>
                                    <button type="submit" className="btn btn-primary">Add</button>
                                </>}
                                {!envList && <div>loading...</div>}
                            </form>
                        </div>
                        <div className="modal-footer">
                            <button type="button" className="btn btn-secondary" data-bs-dismiss="modal">Close</button>
                        </div>
                    </div>
                </div>
            </div>

            <div className="row" style="margin: 10px; height:100vh">
                <div className="col-2 border rounded">
                    <h4 style="padding: 5px 5px">Applications</h4>
                    <hr/>
                    <div className="list-group">
                        {applications && applications.map(elem => (
                            <a href="#" className="list-group-item list-group-item-action" onClick={() => {
                                clearSelected();
                                setSelectedApp([elem.app_id, elem.app_name, elem.env_and_dc]);
                            }}>{elem.app_name}</a>
                        ))}
                    </div>
                    <div>
                        <button type="button" className="btn btn-light border" style="width: 100%;margin-top: 10px;"
                                data-bs-toggle="modal"
                                data-bs-target="#createNewAppModal">New Application
                        </button>
                    </div>
                </div>
                <div className="col-2 border rounded">
                    {selectedApp &&
                        <>
                            <h4 style="padding: 5px 5px">[<strong style="color:red">{selectedApp[1]}</strong>]</h4>
                            <h4 style="padding: 5px 5px">Envs & DCs</h4>
                            <hr/>
                            {
                                applications && applications.map(appItem => {
                                    if (appItem.app_id === selectedApp[0]) {
                                        return (
                                            <div className="accordion" id="env_and_dc_for_application">
                                                {appItem.env_and_dc && appItem.env_and_dc.map(envAndDcItem => (
                                                    <div className="accordion-item">
                                                        <h2 className="accordion-header">
                                                            <button className="accordion-button" type="button"
                                                                    data-bs-toggle="collapse"
                                                                    data-bs-target={'#panelsStayOpen-' + envAndDcItem.env}
                                                                    aria-expanded="true"
                                                                    aria-controls={'panelsStayOpen-' + envAndDcItem.env}>
                                                                {envAndDcItem.env}
                                                            </button>
                                                        </h2>
                                                        <div id={'panelsStayOpen-' + envAndDcItem.env}
                                                             className="accordion-collapse collapse show">
                                                            <div className="accordion-body">
                                                                <div className="list-group">
                                                                    {envAndDcItem.dc_list && envAndDcItem.dc_list.map(dcItem => (
                                                                        <a href="#" onClick={() => {
                                                                            setSelectedEnvAndDc([envAndDcItem.env, dcItem]);
                                                                        }}
                                                                           className="list-group-item list-group-item-action">{dcItem}</a>
                                                                    ))}
                                                                </div>
                                                            </div>
                                                        </div>
                                                    </div>
                                                ))}
                                            </div>
                                        )
                                    }
                                })
                            }
                            <div>
                                <button type="button" className="btn btn-light border"
                                        style="width: 100%;margin-top: 10px;"
                                        data-bs-toggle="modal"
                                        data-bs-target="#addDatacenterModal">Add Datacenter
                                </button>
                            </div>
                        </>
                    }
                </div>
                <div className="col-8 border rounded">
                    {selectedApp && selectedEnvAndDc &&
                        <>
                            <ConfigureListView app_id={selectedApp[0]} app={selectedApp[1]} env={selectedEnvAndDc[0]}
                                               dc={selectedEnvAndDc[1]}/>
                        </>
                    }
                </div>
            </div>
        </div>
    );
}

export function ConfigureListView(props) {

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

    if (!props.app_id || !props.app || !props.env || !props.dc) {
        return (
            <>
                Please select application, environment and datacenter.
            </>
        )
    }

    let [namespaceList, setNamespaceList] = useState(null);
    /*
    [
        {
            "namespace": "namespace_name",
            "configure_list": [
                {
                    "key": "key_name",
                    "configure_id": "123",
                }
            ]
        },
        {
            "namespace": "namespace_name2",
            "configure_list": [
                {
                    "key": "key_name",
                    "configure_id": "123",
                }
            ]
        }
    ]
     */
    let [configureList, setConfigureList] = useState(null);
    let loadConfigureList = function () {
        setConfigureList(null); // clear existing list first
        let app_id = encodeURI(props.app_id);
        let env = encodeURI(props.env);
        let dc = encodeURI(props.dc);
        let url = `/configures/configure_list/${app_id}/${env}/${dc}`
        fetch(ApiEndpoint(url), {
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
                    alert("Get app configure list data error");
                    return
                }
                let dataJson = await res.json();
                console.log("fetch app configure list data:", dataJson.result)
                if (!dataJson.result) {
                    setConfigureList([]);
                } else {
                    setConfigureList(dataJson.result);
                }
            })
            .catch(e => {
                console.log("fetch app configure list data failed:", e)
                //FIXME better error information display
                alert("Get app configure list data error");
            })
    }

    // register form fetch events
    useEffect(e => {
        // trigger fetching namespace list for addConfiguration
        let addConfigureFn = function (event) {
            // clear namespace list
            setNamespaceList(null);
            // fetch new namespace list
            fetch(ApiEndpoint('/configures/namespaces/' + props.app_id), {
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
                        alert("Get app namespaces data error");
                        return
                    }
                    let dataJson = await res.json();
                    console.log("fetch app namespaces data:", dataJson.result)
                    if (!dataJson.result) {
                        setNamespaceList([]);
                    } else {
                        setNamespaceList(dataJson.result);
                    }
                })
                .catch(e => {
                    console.log("fetch app namespaces data failed:", e)
                    //FIXME better error information display
                    alert("Get app namespaces data error");
                })
        }
        document.getElementById('addConfiguration').addEventListener('show.bs.modal', addConfigureFn);
        return () => {
            document.getElementById('addConfiguration').removeEventListener('show.bs.modal', addConfigureFn);
        }
    }, []);
    // reload configure list once any of the parameters changed
    useEffect(e => {
        loadConfigureList();
        return () => {
        }
    }, [props.dc, props.env, props.app]);

    // submit add namespace form
    let submitAddNamespaceForm = function (event, app_id) {
        let nsType = encodeURI(event.target.nstype.value);
        let nsName = encodeURI(event.target.namespace.value);
        let appId = encodeURI(app_id);
        event.target.namespace.value = '';
        fetch(ApiEndpoint('/configures/namespace/' + appId + '/' + nsName + '/' + nsType), {
            headers: {
                'content-type': 'application/json; charset=UTF-8',
                'Authorization': 'Bearer ' + LoginToken(),
            },
            method: 'put',
        })
            .then(async res => {
                console.log("created status:", res.status)
                if (res.status === 401) {
                    setLogout(true);
                    return
                }
                if (res.status !== 200) {
                    console.log("status:", res.status)
                    //FIXME better error information display
                    alert("put new namespace error");
                    return
                }
                // trigger close the modal when complete
                bootstrap.Modal.getInstance(document.getElementById('addNamespace')).hide();
            })
            .catch(e => {
                console.log("put new namespace failed:", e)
                //FIXME better error information display
                alert("put new namespace error");
            })
    }
    // submit add configuration form
    let submitAddConfigurationForm = function (event) {
        let appId = encodeURI(props.app_id);
        let env = encodeURI(props.env);
        let dc = encodeURI(props.dc);
        let nsName = encodeURI(event.target.ns_name.value);
        let key = encodeURI(event.target.key.value);
        let ct = event.target.ct.value;
        let content = event.target.content.value;
        console.log([appId, env, dc, nsName, key, ct, content]);
        fetch(ApiEndpoint('/configures/configure/' + appId + '/' + env + '/' + dc + '/' + nsName + '/' + key), {
            headers: {
                'content-type': 'application/json; charset=UTF-8',
                'Authorization': 'Bearer ' + LoginToken(),
            },
            method: 'post',
            body: JSON.stringify({
                content_type: ct,
                content: content,
            }),
        })
            .then(async res => {
                console.log("created status:", res.status)
                if (res.status === 401) {
                    setLogout(true);
                    return
                }
                if (res.status !== 200) {
                    console.log("status:", res.status)
                    //FIXME better error information display
                    alert("put new configuration error");
                    return
                }
                // trigger reload configure list and close the modal when complete
                loadConfigureList();
                bootstrap.Modal.getInstance(document.getElementById('addConfiguration')).hide();
            })
            .catch(e => {
                console.log("put new configuration failed:", e)
                //FIXME better error information display
                alert("put new configuration error");
            })
    }

    // show usage in code parameters
    // {app, env, dc, ns, key, cfg_id}
    let [selectedConfiguration, setSelectedConfiguration] = useState(null);

    return (
        <>
            <div className="modal fade" id="displayUsageInCodeModel" data-bs-backdrop="static" data-bs-keyboard="false"
                 tabIndex="-1" aria-labelledby="displayUsageInCodeModelLabel" aria-hidden="true">
                <div className="modal-dialog modal-xl">
                    <div className="modal-content">
                        <div className="modal-header">
                            <h1 className="modal-title fs-5" id="displayUsageInCodeModelLabel">Show usage in code</h1>
                            <button type="button" className="btn-close" data-bs-dismiss="modal"
                                    aria-label="Close"></button>
                        </div>
                        <div className="modal-body">
                            {!selectedConfiguration && <div>Loading...</div>}
                            {selectedConfiguration && <>
                                <h5>Current application:[<b>{selectedConfiguration.app}</b>],
                                    environment:[<b>{selectedConfiguration.env}</b>],
                                    datacenter:[<b>{selectedConfiguration.dc}</b>]</h5>
                                <h5>Current namespace:[<b>{selectedConfiguration.ns}</b>],
                                    key:[<b>{selectedConfiguration.key}</b>]</h5>
                                <div className="card" style={"margin: 5px 0px"}>
                                    <div className="card-header">
                                        Go client
                                    </div>
                                    <div className="card-body">
                                        <pre className="card-text" style="font-family: Consolas, Monaco, monospace;">
                                            {`
// Step1. setup selectors when initializing go client
c := NewClient([]string{"http://127.0.0.1:8800"}, ClientOptions{
    SelectorApp:         "${selectedConfiguration.app}",
    SelectorEnvironment: "${selectedConfiguration.env}",
    SelectorDatacenter:  "${selectedConfiguration.dc}",
})

// Step2. register listener information when using the current configuration
atomicContainer, _ := ca.RegisterJsonContainer("${selectedConfiguration.ns}", "${selectedConfiguration.key}", new(ConfigureContainer))
                                            `}
                                        </pre>
                                    </div>
                                </div>

                                <div className="card" style={"margin: 5px 0px"}>
                                    <div className="card-header">
                                        General client (selectors is string formatted)
                                    </div>
                                    <div className="card-body">
                                        <pre className="card-text" style="font-family: Consolas, Monaco, monospace;">
                                            {`
1. Selectors: app=${selectedConfiguration.app},dc=${selectedConfiguration.dc},env=${selectedConfiguration.env}
2. No Optional Selectors set
3. Group: ${selectedConfiguration.ns}
4. Key: ${selectedConfiguration.key}

OnlyAgent flags:
-sel app=${selectedConfiguration.app},dc=${selectedConfiguration.dc},env=${selectedConfiguration.env} -group ${selectedConfiguration.ns} -key ${selectedConfiguration.key}
                                            `}
                                        </pre>
                                    </div>
                                </div>
                            </>}
                        </div>
                        <div className="modal-footer">
                            <button type="button" className="btn btn-secondary" data-bs-dismiss="modal">Close</button>
                        </div>
                    </div>
                </div>
            </div>

            <div className="modal fade" id="addNamespace" data-bs-backdrop="static" data-bs-keyboard="false"
                 tabIndex="-1" aria-labelledby="addNamespaceLabel" aria-hidden="true">
                <div className="modal-dialog modal-lg">
                    <div className="modal-content">
                        <div className="modal-header">
                            <h1 className="modal-title fs-5" id="addNamespaceLabel">Add namespace</h1>
                            <button type="button" className="btn-close" data-bs-dismiss="modal"
                                    aria-label="Close"></button>
                        </div>
                        <div className="modal-body">
                            <form onSubmit={event => {
                                event.preventDefault();
                                submitAddNamespaceForm(event, props.app_id);
                            }}>
                                <h4>Current application:[<strong style="color:red;">{props.app}</strong>]</h4>
                                <h5>Namespace</h5>
                                <input type="text" className="form-control" name="namespace" autoComplete="off"/>
                                <h5>Namespace type</h5>
                                <div className="form-check form-check-inline">
                                    <input className="form-check-input" type="radio" name="nstype" value="application"
                                           id="nstype_application" checked/>
                                    <label className="form-check-label" htmlFor="nstype_application">
                                        Application
                                    </label>
                                </div>
                                <div className="form-check form-check-inline">
                                    <input className="form-check-input" type="radio" name="nstype" value="public"
                                           id="nstype_public"/>
                                    <label className="form-check-label" htmlFor="nstype_public">
                                        Public
                                    </label>
                                </div>
                                <hr/>
                                <button type="submit" className="btn btn-primary">Add</button>
                            </form>
                        </div>
                        <div className="modal-footer">
                            <button type="button" className="btn btn-secondary" data-bs-dismiss="modal">Close</button>
                        </div>
                    </div>
                </div>
            </div>

            <div className="modal fade" id="addConfiguration" data-bs-backdrop="static" data-bs-keyboard="false"
                 tabIndex="-1" aria-labelledby="addConfigurationLabel" aria-hidden="true">
                <div className="modal-dialog modal-xl">
                    <div className="modal-content">
                        <div className="modal-header">
                            <h1 className="modal-title fs-5" id="addConfigurationLabel">Add configuration</h1>
                            <button type="button" className="btn-close" data-bs-dismiss="modal"
                                    aria-label="Close"></button>
                        </div>
                        <div className="modal-body">
                            <form onSubmit={event => {
                                event.preventDefault();
                                submitAddConfigurationForm(event);
                            }}>
                                {!namespaceList && <div>Loading...</div>}
                                {namespaceList && <>
                                    <h5>Namespace</h5>
                                    <select className="form-select" name="ns_name">
                                        {namespaceList.map(elem => (
                                            <option value={elem}>{elem}</option>
                                        ))}
                                    </select>
                                    <h5>Key</h5>
                                    <input type="text" className="form-control" name="key" autoComplete="off"/>
                                    <h5>Content type</h5>
                                    <div>
                                        <div className="form-check form-check-inline">
                                            <input className="form-check-input" type="radio" name="ct"
                                                   value="general"
                                                   id="ct_general" checked/>
                                            <label className="form-check-label" htmlFor="ct_general">
                                                General Text
                                            </label>
                                        </div>
                                    </div>
                                    <h5>Configure content</h5>
                                    <textarea className="form-control" name="content" rows="20"
                                              style="font-family: Consolas, Monaco, monospace;"></textarea>
                                    <hr/>
                                    <button type="submit" className="btn btn-primary">Add</button>
                                </>}
                            </form>
                        </div>
                        <div className="modal-footer">
                            <button type="button" className="btn btn-secondary" data-bs-dismiss="modal">Close</button>
                        </div>
                    </div>
                </div>
            </div>

            <EditConfigureModal selected={selectedConfiguration} modal_id={'editConfigurationEach'}/>

            <h4 style="margin: 5px 5px">Configurations</h4>
            <h4 style="margin: 5px 5px">Current Application:[<strong style="color:red">{props.app}</strong>]</h4>
            <h4 style="margin: 5px 5px">Current Environment:[<strong style="color:red">{props.env}</strong>],
                Datacenter:[<strong style="color:red">{props.dc}</strong>]</h4>
            <div className="btn-group" role="group">
                <button type="button" className="btn btn-outline-primary" data-bs-toggle="modal"
                        data-bs-target="#addConfiguration">Add Configure
                </button>
                <button type="button" className="btn btn-outline-primary" data-bs-toggle="modal"
                        data-bs-target="#addNamespace">Add Namespace
                </button>
                <button type="button" className="btn btn-outline-primary" disabled>Link Namespace</button>
            </div>
            <hr/>
            {configureList && configureList.map(elem => (
                <div className="card" style="margin-bottom: 5px;">
                    <div className="card-header">
                        Namespace: <strong>{elem.namespace}</strong>
                    </div>
                    <div className="card-body">
                        <ul className="list-group list-group-flush">
                            {elem.configure_list && elem.configure_list.map(configElem => (
                                    <li className="list-group-item">
                                        <div><h5>Key: <strong>{configElem.key}</strong></h5></div>
                                        <div className="btn-group" role="group">
                                            <button type="button" className="btn btn-outline-primary" data-bs-toggle="modal"
                                                    data-bs-target={"#" + 'editConfigurationEach'} onClick={() => {
                                                setSelectedConfiguration(null);
                                                setSelectedConfiguration({
                                                    app: props.app,
                                                    env: props.env,
                                                    dc: props.dc,
                                                    ns: elem.namespace,
                                                    key: configElem.key,
                                                    cfg_id: configElem.configure_id,
                                                });
                                            }}>Edit
                                            </button>
                                            <button type="button" className="btn btn-outline-primary" data-bs-toggle="modal"
                                                    data-bs-target="#displayUsageInCodeModel" onClick={() => {
                                                setSelectedConfiguration(null);
                                                setSelectedConfiguration({
                                                    app: props.app,
                                                    env: props.env,
                                                    dc: props.dc,
                                                    ns: elem.namespace,
                                                    key: configElem.key,
                                                    cfg_id: configElem.configure_id,
                                                });
                                            }}>View usage in code
                                            </button>
                                            <button type="button" className="btn btn-outline-primary"
                                                    disabled>History
                                            </button>
                                            <button type="button" className="btn btn-outline-danger"
                                                    disabled>Rollback
                                            </button>
                                        </div>
                                    </li>
                                )
                            )}
                        </ul>
                    </div>
                </div>
            ))}
            {!configureList && <div>
                Loading...
            </div>}
        </>
    )
}

export function EditConfigureModal(props) {
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

    let [prevData, setPrevData] = useState(null);
    useEffect(() => {
        // trigger fetching previous configuration data
        let fetchPrevConfigureFn = function (event) {
            // clear namespace list
            setPrevData(null);
            // fetch new namespace list
            let url = `/configures/configure/${props.selected.cfg_id}`
            fetch(ApiEndpoint(url), {
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
                        alert("Get previous configure data error");
                        return
                    }
                    let dataJson = await res.json();
                    console.log("fetch previous configure data:", dataJson.result)
                    if (!dataJson.result) {
                        setPrevData({});
                    } else {
                        setPrevData(dataJson.result);
                    }
                })
                .catch(e => {
                    console.log("fetch previous configure data failed:", e)
                    //FIXME better error information display
                    alert("Get previous configure data error");
                })
        }
        //FIXME This method will be triggered everytime 'selected' item changed. But it is also incompatible with
        // 'model show' event in bootstrap when using document.addEventListener, data-bs-toggle and onClick function on button.
        if (props.selected) {
            fetchPrevConfigureFn();
        }
    }, [props.selected]);

    // submit editConfiguration modal form
    let submitEditConfigurationForm = function (event) {
        let cfgId = props.selected.cfg_id;
        let contentType = event.target.ct.value;
        let content = event.target.content.value;
        console.log("edit request:", cfgId, contentType, content);
        let url = `/configures/configure/${cfgId}`
        fetch(ApiEndpoint(url), {
            headers: {
                'content-type': 'application/json; charset=UTF-8',
                'Authorization': 'Bearer ' + LoginToken(),
            },
            method: 'put',
            body: JSON.stringify({
                ct: contentType,
                content: content,
            })
        })
            .then(async res => {
                console.log("created status:", res.status)
                if (res.status === 401) {
                    setLogout(true);
                    return
                }
                if (res.status !== 200) {
                    console.log("status:", res.status)
                    //FIXME better error information display
                    alert("put edit configuration error");
                    return
                }
                // trigger close the modal when complete
                bootstrap.Modal.getInstance(document.getElementById(props.modal_id)).hide();
            })
            .catch(e => {
                console.log("put edit configuration failed:", e)
                //FIXME better error information display
                alert("put edit configuration error");
            })
    }

    return (
        <>
            <div className="modal fade" id={props.modal_id} data-bs-backdrop="static" data-bs-keyboard="false"
                 tabIndex="-1" aria-labelledby={props.modal_id + 'Label'} aria-hidden="true">
                <div className="modal-dialog modal-xl">
                    <div className="modal-content">
                        <div className="modal-header">
                            <h1 className="modal-title fs-5" id={props.modal_id + 'Label'}>Edit configuration</h1>
                            <button type="button" className="btn-close" data-bs-dismiss="modal"
                                    aria-label="Close"></button>
                        </div>
                        <div className="modal-body">
                            <form onSubmit={event => {
                                event.preventDefault();
                                submitEditConfigurationForm(event);
                            }}>
                                {!prevData && <div>Loading...</div>}
                                {prevData && <>
                                    <h5>Namespace</h5>
                                    <input type="text" className="form-control" name="ns_name" autoComplete="off"
                                           disabled value={prevData.cfg_ns}/>
                                    <h5>Key</h5>
                                    <input type="text" className="form-control" name="key" autoComplete="off"
                                           disabled value={prevData.cfg_key}/>
                                    <h5>Content type</h5>
                                    <div>
                                        <div className="form-check form-check-inline">
                                            <input className="form-check-input" type="radio" name="ct"
                                                   value="general"
                                                   id="ct_general" checked={prevData.cfg_ct === "general"}/>
                                            <label className="form-check-label" htmlFor="ct_general">
                                                General Text
                                            </label>
                                        </div>
                                    </div>
                                    <h5>Configure content</h5>
                                    <textarea className="form-control" name="content" rows="20"
                                              style="font-family: Consolas, Monaco, monospace;">{prevData.cfg_content}</textarea>
                                    <hr/>
                                    <button type="submit" className="btn btn-primary">Add</button>
                                </>}
                            </form>
                        </div>
                        <div className="modal-footer">
                            <button type="button" className="btn btn-secondary" data-bs-dismiss="modal">Close</button>
                        </div>
                    </div>
                </div>
            </div>
        </>
    )
}
