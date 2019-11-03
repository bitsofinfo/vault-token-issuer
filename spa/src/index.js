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
          this.setState({renewable: Boolean(event.target.value)});

        } else if (event.target.id === 'period') {
          this.setState({period: event.target.value});
        }
    }

    postTokenCreateOrphan() {

        axios.post('https://localhost:8443/token/create-orphan', 
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

        <div class="tokenForm">

          <div class="title">Vault: Create orphan token</div>

          <form onSubmit={this.handleSubmit}>

          <div class="section">
            <div class="sectionTitle">Authentication</div>

            <div class="field">
              <input type="text" placeholder="Username" id="username" value={this.state.username} onChange={this.handleChange} required/>
              <span class="help-text"></span>
            </div>

            <div class="field">
                <input type="password" placeholder="Password" id="password" value={this.state.password} onChange={this.handleChange} required/>
                <span class="help-text"></span>
            </div>
          </div>

          <div class="section">

            <div class="sectionTitle">Token options</div>

            <div class="field">
                <textarea id="policies" required placeholder="Enter one or more vault token policy names; separated by new lines" value={this.state.policies.join("\n")} onChange={this.handleChange} />
                <span class="help-text"></span>
            </div>

            <div class="select-grid">

                <div class="select-label">Renewable?</div>
                <div class="select-control">
                  <select id="renewable" value={this.state.renewable.toString()} onChange={this.handleChange}>
                    <option value="true">Yes</option>
                    <option value="false">No</option>
                  </select>
                </div>

                <div class="select-label">Period</div>
                <div class="select-control">
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

          <div class="controls">
            <input id="submit" type="submit" value="create token" />
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
        success: (result.status === 200 ? true : false),
        token: (result.data ? result.data.token  : null),
        code: (result.data ? result.data.code  : null),
        msg: (result.data ? result.data.msg  : null) }
      );
    }

    render() {
      return (
        <div id="results" class="section">
          Status: {this.state.status}
          Token: {this.state.token}
          Code: {this.state.code}
          Msg: {this.state.msg}
        </div>
      )
    }
  }

  
  
  // ========================================
  
  ReactDOM.render(
    <CreateTokenForm />,
    document.getElementById('root')
  );
  