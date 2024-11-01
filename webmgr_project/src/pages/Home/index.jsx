import './style.css';
import {IsLogin} from "../../components/loginstatus.jsx";
import Redirect from "../../components/redirect.jsx";
import {OnlyConfigHome} from "../OnlyConfig/configure.jsx";

export function Home() {
    if (!IsLogin()) {
        return (
            <Redirect to='/login'/>
        );
    }

    return (
        <div>
            <OnlyConfigHome/>
        </div>
    );
}
