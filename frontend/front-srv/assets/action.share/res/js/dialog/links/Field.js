/*
 * Copyright 2007-2017 Charles du Jeu - Abstrium SAS <team (at) pyd.io>
 * This file is part of Pydio.
 *
 * Pydio is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */
import React from 'react';
import Pydio from 'pydio'
import PropTypes from 'prop-types';
import ShareContextConsumer from '../ShareContextConsumer'
import TargetedUsers from './TargetedUsers'
import {TextField, Paper} from 'material-ui'
import QRCode from 'qrcode.react'
import Clipboard from 'clipboard'
import ActionButton from '../main/ActionButton'
import PathUtils from 'pydio/util/path'
import LangUtils from 'pydio/util/lang'
import {muiThemeable} from 'material-ui/styles'
import ShareHelper from '../main/ShareHelper'
const {ModernStyles} = Pydio.requireLib('hoc')
const {Tooltip} = Pydio.requireLib("boot");

import LinkModel from './LinkModel'

class PublicLinkField extends React.Component {
    static propTypes = {
        linkModel: PropTypes.instanceOf(LinkModel),
        editAllowed: PropTypes.bool,
        onChange: PropTypes.func,
        showMailer:PropTypes.func
    };

    state = {editLink: false, copyMessage:'', showQRCode: false};

    toggleEditMode = () => {
        const {linkModel, pydio} = this.props;
        if(this.state.editLink && this.state.customLink){
            const auth = ShareHelper.getAuthorizations();
            if(auth.hash_min_length && this.state.customLink.length < auth.hash_min_length){
                pydio.UI.displayMessage('ERROR', this.props.getMessage('223').replace('%s', auth.hash_min_length));
                return;
            }
            linkModel.setCustomLink(this.state.customLink);
            linkModel.save();
        }
        this.setState({editLink: !this.state.editLink, customLink: undefined});
    };

    changeLink = (event) => {
        let value = event.target.value;
        value = LangUtils.computeStringSlug(value);
        this.setState({customLink: value});
    };

    clearCopyMessage = () => {
        global.setTimeout(function(){
            this.setState({copyMessage:''});
        }.bind(this), 5000);
    };

    attachClipboard = () => {
        const {linkModel, pydio} = this.props;
        this.detachClipboard();
        if(this.refs['copy-button']){
            this._clip = new Clipboard(this.refs['copy-button'], {
                text: function(trigger) {
                    return ShareHelper.buildPublicUrl(pydio, linkModel.getLink());
                }.bind(this)
            });
            this._clip.on('success', function(){
                this.setState({copyMessage:this.props.getMessage('192')}, this.clearCopyMessage);
            }.bind(this));
            this._clip.on('error', function(){
                let copyMessage;
                if( global.navigator.platform.indexOf("Mac") === 0 ){
                    copyMessage = this.props.getMessage('144');
                }else{
                    copyMessage = this.props.getMessage('143');
                }
                this.refs['public-link-field'].focus();
                this.setState({copyMessage:copyMessage}, this.clearCopyMessage);
            }.bind(this));
        }
    };

    detachClipboard = () => {
        if(this._clip){
            this._clip.destroy();
        }
    };

    componentDidUpdate(prevProps, prevState) {
        this.attachClipboard();
    }

    componentDidMount() {
        this.attachClipboard();
    }

    componentWillUnmount() {
        this.detachClipboard();
    }

    openMailer = () => {
        this.props.showMailer(this.props.linkModel);
    };

    toggleQRCode = () => {
        this.setState({showQRCode:!this.state.showQRCode});
    };

    confirmDisable() {
        const {onDisableLink} = this.props;
        if (confirm(this.props.getMessage('dialog.link.confirm.remove'))){
            onDisableLink()
        }
    }

    render() {
        const {linkModel, pydio} = this.props;
        const publicLink = ShareHelper.buildPublicUrl(pydio, linkModel.getLink());
        const auth = ShareHelper.getAuthorizations();
        const editAllowed = this.props.editAllowed && auth.editable_hash && !this.props.isReadonly() && linkModel.isEditable();
        if(this.state.editLink && editAllowed){
            return (
                <div>
                    <div style={{display:'flex', alignItems:'center', backgroundColor: 'rgb(246, 246, 248)', padding: 6, borderRadius: 2}}>
                        <span style={{fontSize:16, color:'rgba(0,0,0,0.4)', display: 'inline-block', maxWidth: 160, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis'}}>{PathUtils.getDirname(publicLink) + '/ '}</span>
                        <TextField style={{flex:1, marginRight: 10, marginLeft: 10}} onChange={this.changeLink} value={this.state.customLink !== undefined ? this.state.customLink : linkModel.getLink().LinkHash}/>
                        <ActionButton mdiIcon="check" callback={this.toggleEditMode} />
                    </div>
                    <div style={{textAlign:'center', fontSize:13, color:'rgba(0,0,0,0.43)', paddingTop: 16}}>{this.props.getMessage('194')}</div>
                </div>
            );
        }else{
            const {copyMessage, linkTooltip} = this.state;
            const setHtml = function(){
                return {__html:this.state.copyMessage};
            }.bind(this);
            let actionLinks = [], qrCode;
            const {muiTheme} = this.props;

            const copyButton = (
                <div
                    key={"copy"}
                    ref="copy-button"
                    style={{position: 'absolute', right: 0, bottom: 7, width:30, height:30, padding:4, backgroundColor:'#f6f6f8', fontSize:16, cursor:'pointer'}}
                    onMouseOver={()=>{this.setState({linkTooltip:true})}}
                    onMouseOut={()=>{this.setState({linkTooltip:false})}}
                >
                    <Tooltip
                        label={copyMessage ? copyMessage : this.props.getMessage('191')}
                        horizontalPosition={"left"}
                        verticalPosition={"bottom"}
                        show={linkTooltip}
                    />
                    <span className="copy-link-button mdi mdi-content-copy" style={{color: muiTheme.palette.primary1Color}}/>
                </div>
            )

            if(this.props.showMailer){
                actionLinks.push(<ActionButton key="outline" callback={this.openMailer} mdiIcon="email-outline" messageId="45"/>);
            }
            if(editAllowed){
                actionLinks.push(<ActionButton key="pencil" callback={this.toggleEditMode} mdiIcon="pencil" messageId={"193"}/>);
            }
            if(ShareHelper.qrcodeEnabled()){
                actionLinks.push(<ActionButton key="qrcode" callback={this.toggleQRCode} mdiIcon="qrcode" messageId={'94'}/>);
            }
            if(this.props.onDisableLink) {
                actionLinks.push(<ActionButton key="delete" destructive={true} callback={() => this.confirmDisable()} mdiIcon="link-off" messageId="link.disable"/>);
            }
            if(actionLinks.length){
                actionLinks = (
                    <div style={{display:'flex', marginTop:10}}><span style={{flex:1}}/>{actionLinks}<span style={{flex:1}}/></div>
                ) ;
            }else{
                actionLinks = null;
            }
            if(this.state.showQRCode){
                qrCode = <Paper zDepth={1} style={{width:120, paddingTop:10, overflow:'hidden', margin:'0 auto', height:120, textAlign:'center'}}><QRCode size={100} value={publicLink} level="Q"/></Paper>;
            } else {
                qrCode = <Paper zDepth={0} style={{width:120, overflow:'hidden', margin:'0 auto', height:0, textAlign:'center'}}></Paper>
            }
            return (
                <Paper zDepth={0} rounded={false} className="public-link-container">
                    <div style={{marginTop:-8, position:'relative'}}>
                        <TextField
                            floatingLabelText={this.props.getMessage("link.floatingLabel")}
                            floatingLabelFixed={true}
                            type="text"
                            name="Link"
                            ref="public-link-field"
                            value={publicLink}
                            onFocus={e => {e.target.select()}}
                            fullWidth={true}
                            {...ModernStyles.textFieldV2}
                        />
                        {copyButton}
                    </div>
                    {false && this.props.linkData.target_users && <TargetedUsers {...this.props}/>}
                    {actionLinks}
                    {qrCode}
                </Paper>
            );
        }
    }
}

PublicLinkField = muiThemeable()(PublicLinkField);
PublicLinkField = ShareContextConsumer(PublicLinkField)
export {PublicLinkField as default};