export function OnlyConfigNavBar() {
    return (
        <nav className="navbar bg-body-tertiary">
            <div className="container-fluid">
                <a className="navbar-brand" href="/">OnlyConfig Web Manager</a>
                <div className="d-flex btn-group">
                    <a className='btn btn-outline-success' href='/'>Home</a>
                    <a className='btn btn-outline-primary' href='/change_password'>Change password</a>
                    <a className='btn btn-outline-primary' href='/env_and_dc'>Manage environment and datacenter</a>
                    <a className='btn btn-outline-primary' href='/org_mgr'>Manage organization</a>
                    <a className='btn btn-outline-danger' href='/logout'>Logout</a>
                </div>
            </div>
        </nav>
    );
}
