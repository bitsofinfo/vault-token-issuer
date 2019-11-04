import React from 'react';
import ReactDOM from 'react-dom';
import './index.css';

const axios = require('axios');

class CreateTokenForm extends React.Component {

    constructor(props) {
        super(props);
        this.state = {username:'', 
                      password:'', 
                      period: "5m", 
                      renewable: true, 
                      policies:[], 
                      result: null};

        this.createTokenResult = React.createRef();
      
        this.handleChange = this.handleChange.bind(this);
        this.handleSubmit = this.handleSubmit.bind(this);
    }

    handleChange(event) {
        if (event.target.id === 'username') {
          this.setState({username: event.target.value});

        } else if (event.target.id === 'password') {
          this.setState({password: event.target.value});

        } else if (event.target.id === 'policies') {
          this.setState({policies: event.target.value.split("\n")});

        } else if (event.target.id === 'renewable') {
          this.setState({renewable: (event.target.value === 'true' ? true : false)});

        } else if (event.target.id === 'period') {
          this.setState({period: event.target.value});
        }
    }

    postTokenCreateOrphan() {

        axios.post((process.env.REACT_APP_VAULT_TOKEN_ISSUER_ROOT_URL+'/token/create-orphan'), 
          {
            renewable: this.state.renewable,
            period: this.state.period,
            policies: this.state.policies.filter(x => x.trim() !== '')
          },{
            headers: {'Content-Type': 'application/json'},
            auth: {
              username: this.state.username,
              password: this.state.password
            }
        })
        
        .then(function (response) {
          console.log(response)
          this.setState({result:{
            status: response.status,
            statusText: response.statusText,
            data: response.data
          }});
          this.createTokenResult.current.update(this.state.result);
        }.bind(this)) 

        .catch(function (error) {
          console.log(error)

          if (error.response) {
            this.setState({ result: {
              status: error.response.status,
              statusText: error.response.statusText,
              data: error.response.data
            }});
          } else {
            this.setState({ result: {
              status: error.toString(),
              statusText: error.toString(),
              data: null
            }});
          }
         
          this.createTokenResult.current.update(this.state.result);
        }.bind(this));
    }


    handleSubmit(event) {
        this.postTokenCreateOrphan()
        event.preventDefault();
    }

    render() {
      return (

        <div className="tokenForm">

          <div className="title">Vault: Create orphan token</div>

          <form onSubmit={this.handleSubmit}>

          <div className="section">
            <div className="sectionTitle">Authentication</div>

            <div className="field">
              <input type="text" placeholder="Username" id="username" value={this.state.username} onChange={this.handleChange} required/>
              <span className="help-text"></span>
            </div>

            <div className="field">
                <input type="password" placeholder="Password" id="password" value={this.state.password} onChange={this.handleChange} required/>
                <span className="help-text"></span>
            </div>
          </div>

          <div className="section">

            <div className="sectionTitle">Token options</div>

            <div className="field">
                <textarea id="policies" required placeholder="Enter one or more vault token policy names; separated by new lines" value={this.state.policies.join("\n")} onChange={this.handleChange} />
                <span className="help-text"></span>
            </div>

            <div className="select-grid">

                <div className="select-label">Renewable?</div>
                <div className="select-control">
                  <select id="renewable" onChange={this.handleChange}>
                    <option value="true" selected={this.state.renewable}>Yes</option>
                    <option value="false" selected={!this.state.renewable}>No</option>
                  </select>
                </div>

                <div className="select-label">Period</div>
                <div className="select-control">
                  <select id="period" value={this.state.period} onChange={this.handleChange}>
                    <option value="1m">1m</option>
                    <option value="5m">5m</option>
                    <option value="10m">10m</option>
                    <option value="20m">20m</option>
                    <option value="30m">30m</option>
                    <option value="60m">60m</option>
                  </select>
                </div>  

            </div>

          </div>

          { this.state.result != null ? <CreateTokenResults ref={this.createTokenResult}/> : null }

          <div className="controls">
            <input className="button" id="submit" type="submit" value="Generate orphan token" />
          </div>

        </form>

      </div>
      );
    }
  }

  class CreateTokenResults extends React.Component {

    constructor(props) {
      super(props);
      this.state = { }
      this.update = this.update.bind(this)
    }

    update(result) {
      this.setState( { 
        status: result.status, 
        statusText: result.statusText, 
        data: result.data,
        json: JSON.stringify(result.data,undefined, 2),
        success: (result.status === 200 ? true : false),
        token: (result.data ? result.data.token  : null),
        code: (result.data ? result.data.code  : null),
        msg: (result.data ? result.data.msg  : null) }
      );
    }

    render() {
      return (
        <div className={this.state.status === 200 ? "results section section-ok" : "results section section-error"}>
          
          <div className="sectionTitle">/token/create-orphan results</div>

          <div className="result-grid">
            <label className="result-label">Status:</label><label className="result-data">{this.state.status} {this.state.statusText}</label>
            <label className="result-label">Token:</label><label className="result-data token">{this.state.token}</label>
            <label className="result-label">Code:</label><label className="result-data">{this.state.code}</label>
            <label className="result-label">Msg:</label><label className="result-data">{this.state.msg}</label>
          </div>
        </div>
      )
    }
  }

  
  
  // ========================================
  
  ReactDOM.render(
    <CreateTokenForm />,
    document.getElementById('root')
  );
  