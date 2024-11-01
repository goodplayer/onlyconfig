import {render} from 'preact';
import {LocationProvider, Route, Router} from 'preact-iso';
import {Home} from './pages/Home/index.jsx';
import {ChangePassword, LoginPage, Logout, RegisterPage} from './pages/Home/login.jsx';
import {NotFound} from './pages/_404.jsx';
import './style.css';
import {EnvAndDc} from "./pages/OnlyConfig/env_dc.jsx";
import {OrgManagement} from "./pages/OnlyConfig/org.jsx";

export function App() {
    return (
        <LocationProvider>
            <main>
                <Router>
                    <Route path="/" component={Home}/>
                    <Route path="/login" component={LoginPage}/>
                    <Route path="/logout" component={Logout}/>
                    <Route path={"/register"} component={RegisterPage}/>
                    <Route path={"/change_password"} component={ChangePassword}/>

                    <Route path="/env_and_dc" component={EnvAndDc}/>
                    <Route path="/org_mgr" component={OrgManagement}/>

                    <Route default component={NotFound}/>
                </Router>
            </main>
        </LocationProvider>
    );
}

render(<App/>, document.getElementById('app'));
