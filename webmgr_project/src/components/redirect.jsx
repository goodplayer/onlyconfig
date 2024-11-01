import {Component} from 'preact';

export default class Redirect extends Component {
    componentWillMount() {
        location = this.props.to;
    }

    render() {
        return null;
    }
}
