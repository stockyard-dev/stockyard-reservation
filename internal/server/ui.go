package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Reservation</title>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital,wght@0,400;0,700;1,400&family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--blue:#5b8dd9;--mono:'JetBrains Mono',monospace;--serif:'Libre Baskerville',serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--serif);line-height:1.6}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-family:var(--mono);font-size:.9rem;letter-spacing:2px}.hdr h1 span{color:var(--rust)}
.main{padding:1.5rem;max-width:1100px;margin:0 auto}
.stats{display:grid;grid-template-columns:repeat(3,1fr);gap:.5rem;margin-bottom:1rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.6rem;text-align:center;font-family:var(--mono)}.st-v{font-size:1.3rem;font-weight:700}.st-l{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.15rem}
.toolbar{display:flex;gap:.5rem;margin-bottom:1rem;align-items:center;flex-wrap:wrap}
.search{flex:1;min-width:180px;padding:.4rem .6rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}.search:focus{outline:none;border-color:var(--leather)}
.filter-sel{padding:.4rem .5rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.65rem}
.tbl-wrap{overflow-x:auto}.tbl{width:100%;border-collapse:collapse;font-family:var(--mono);font-size:.65rem}
.tbl th{text-align:left;padding:.5rem .4rem;border-bottom:2px solid var(--bg3);color:var(--cm);font-size:.55rem;text-transform:uppercase;letter-spacing:1px;white-space:nowrap}
.tbl td{padding:.45rem .4rem;border-bottom:1px solid var(--bg3);max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.tbl tr:hover td{background:var(--bg2);cursor:pointer}
.btn{font-family:var(--mono);font-size:.6rem;padding:.25rem .5rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);transition:all .2s}.btn:hover{border-color:var(--leather);color:var(--cream)}.btn-p{background:var(--rust);border-color:var(--rust);color:#fff}.btn-d{color:var(--red)}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.65);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:520px;max-width:92vw;max-height:90vh;overflow-y:auto}
.modal h2{font-family:var(--mono);font-size:.8rem;margin-bottom:1rem;color:var(--rust);letter-spacing:1px}
.fr{margin-bottom:.6rem}.fr label{display:block;font-family:var(--mono);font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.4rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.fr input:focus,.fr select:focus,.fr textarea:focus{outline:none;border-color:var(--leather)}
.row2{display:grid;grid-template-columns:1fr 1fr;gap:.5rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:1rem}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.85rem}
.count-label{font-family:var(--mono);font-size:.6rem;color:var(--cm);margin-bottom:.5rem}
@media(max-width:600px){.row2{grid-template-columns:1fr}.toolbar{flex-direction:column}.search{min-width:100%}.trial-bar{flex-direction:column;align-items:stretch}.trial-bar input.key-input{width:100%}}
.trial-bar{display:none;background:linear-gradient(90deg,#3a2419,#2e1c14);border-bottom:2px solid var(--rust);padding:.7rem 1.5rem;font-family:var(--mono);font-size:.68rem;color:var(--cream);align-items:center;gap:1rem;flex-wrap:wrap}
.trial-bar.show{display:flex}
.trial-bar-msg{flex:1;min-width:240px;line-height:1.5}
.trial-bar-msg strong{color:var(--rust);text-transform:uppercase;letter-spacing:1px;font-size:.6rem;display:block;margin-bottom:.15rem}
.trial-bar-actions{display:flex;gap:.5rem;align-items:center;flex-wrap:wrap}
.trial-bar a.btn-trial{background:var(--rust);color:#fff;padding:.4rem .8rem;text-decoration:none;font-size:.65rem;text-transform:uppercase;letter-spacing:1px;font-weight:700;border:1px solid var(--rust);transition:all .2s}
.trial-bar a.btn-trial:hover{background:#f08545;border-color:#f08545}
.trial-bar-divider{color:var(--cm);font-size:.6rem}
.trial-bar input.key-input{padding:.4rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.6rem;width:200px}
.trial-bar input.key-input:focus{outline:none;border-color:var(--rust)}
.trial-bar button.btn-activate{padding:.4rem .7rem;background:var(--bg2);color:var(--cream);border:1px solid var(--leather);font-family:var(--mono);font-size:.6rem;cursor:pointer;text-transform:uppercase;letter-spacing:1px}
.trial-bar button.btn-activate:hover{background:var(--bg3)}
.trial-bar button.btn-activate:disabled{opacity:.5;cursor:wait}
.trial-msg{font-size:.6rem;color:var(--cm);margin-left:.5rem}
.trial-msg.error{color:#e74c3c}
.trial-msg.success{color:#4ade80}
.btn-disabled-trial{opacity:.45;cursor:not-allowed!important}
</style></head><body>
<div class="trial-bar" id="trial-bar">
<div class="trial-bar-msg">
<strong>License Required</strong>
<span id="trial-bar-text">You can view your existing data, but writes are locked until you start a 14-day free trial or activate a license key.</span>
</div>
<div class="trial-bar-actions">
<a class="btn-trial" href="https://stockyard.dev/" target="_blank" rel="noopener">Start 14-Day Trial</a>
<span class="trial-bar-divider">or</span>
<input type="text" class="key-input" id="trial-key-input" placeholder="SY-..." autocomplete="off" spellcheck="false">
<button class="btn-activate" id="trial-activate-btn" onclick="activateLicense()">Activate</button>
<span class="trial-msg" id="trial-msg"></span>
</div>
</div>
<div class="hdr"><h1><span>&#9670;</span> RESERVATION</h1><button class="btn btn-p" id="add-btn">+ Add</button></div>
<div class="main">
<div class="stats" id="stats-reservations"></div>
<div class="toolbar">
<input class="search" id="search-reservations" placeholder="Search..." oninput="renderRes('reservations')"><select class="filter-sel" id="filter-reservations-status" onchange="renderRes('reservations')"><option value="">All Status</option><option value="Confirmed">Confirmed</option><option value="Seated">Seated</option><option value="Completed">Completed</option><option value="No Show">No Show</option><option value="Cancelled">Cancelled</option></select></div>
<div class="count-label" id="count-reservations"></div>
<div class="tbl-wrap"><table class="tbl"><thead><tr><th>Guest Name</th><th>Phone</th><th>Email</th><th>Party Size</th><th>Date</th><th>Time</th><th>Table</th><th>Status</th><th>Notes</th><th></th></tr></thead><tbody id="tbody-reservations"></tbody></table></div></div><div class="modal-bg" id="mbg" onclick="if(event.target===this)closeModal()"><div class="modal" id="mdl"></div></div><script>var A="/api",data={},editId=null,activeRes="reservations";var resCfg={"reservations":{d:"Reservations",f:[{n:"guest_name",l:"Guest Name",t:"text",r:true,o:[]},{n:"guest_phone",l:"Phone",t:"phone",r:false,o:[]},{n:"guest_email",l:"Email",t:"email",r:false,o:[]},{n:"party_size",l:"Party Size",t:"integer",r:true,o:[]},{n:"date",l:"Date",t:"date",r:true,o:[]},{n:"time",l:"Time",t:"text",r:true,o:[]},{n:"table",l:"Table",t:"text",r:false,o:[]},{n:"status",l:"Status",t:"select",r:false,o:["Confirmed","Seated","Completed","No Show","Cancelled"]},{n:"notes",l:"Notes",t:"textarea",r:false,o:[]}]},};

async function loadAll(){
for(var k in resCfg){
var r=await fetch(A+'/'+k).then(function(x){return x.json()});
var items=r[k]||[];
try{
var extras=await fetch(A+'/extras/'+k).then(function(x){return x.json()});
items.forEach(function(it){if(extras[it.id]){Object.keys(extras[it.id]).forEach(function(key){if(it[key]===undefined)it[key]=extras[it.id][key]})}});
}catch(e){}
data[k]=items;
}
renderStats();renderRes(activeRes);
}

function renderStats(){
var r=activeRes,items=data[r]||[];
var total=items.length;
var now=new Date(),weekAgo=new Date(now-7*86400000),monthAgo=new Date(now-30*86400000);
var thisWeek=items.filter(function(x){return new Date(x.created_at)>=weekAgo}).length;
var thisMonth=items.filter(function(x){return new Date(x.created_at)>=monthAgo}).length;
document.getElementById('stats-'+r).innerHTML=[
{l:'Total',v:total},{l:'This Week',v:thisWeek},{l:'This Month',v:thisMonth}
].map(function(x){return '<div class="st"><div class="st-v">'+x.v+'</div><div class="st-l">'+x.l+'</div></div>'}).join('');
}

function renderRes(r){
activeRes=r;renderStats();
var cfg=resCfg[r],items=data[r]||[];
var q=(document.getElementById('search-'+r)||{}).value||'';q=q.toLowerCase();
if(q)items=items.filter(function(x){return cfg.f.some(function(f){var v=x[f.n];return v&&String(v).toLowerCase().includes(q)})});
cfg.f.forEach(function(f){if(f.t==='select'){var sel=document.getElementById('filter-'+r+'-'+f.n);if(sel&&sel.value)items=items.filter(function(x){return x[f.n]===sel.value})}});
document.getElementById('count-'+r).textContent=items.length+' record'+(items.length!==1?'s':'');
var tbody=document.getElementById('tbody-'+r);
if(!items.length){var emsg=window._emptyMsg||'No records found.';tbody.innerHTML='<tr><td colspan="'+(cfg.f.length+1)+'" class="empty">'+emsg+'</td></tr>';return}
var h='';items.forEach(function(item){
var rowClick=window._trialRequired?'showTrialNudge()':'openEdit(\\''+r+'\\',\\''+item.id+'\\')';
h+='<tr onclick="'+rowClick+'">';
cfg.f.forEach(function(f){
var v=item[f.n];if(v===undefined||v===null)v='';
if(f.t==='checkbox')v=v?'Yes':'No';
h+='<td>'+esc(String(v))+'</td>';
});
if(window._trialRequired){
h+='<td></td>';
}else{
h+='<td><button class="btn btn-d" onclick="event.stopPropagation();del(\''+r+'\',\''+item.id+'\')">&#10005;</button></td>';
}
h+='</tr>';
});
tbody.innerHTML=h;
}

function formHTML(r,item){
var cfg=resCfg[r],isEdit=!!item;
var i=item||{};
var h='<h2>'+(isEdit?'EDIT':'NEW')+' '+cfg.d.toUpperCase()+'</h2>';
cfg.f.forEach(function(f,idx){
var v=i[f.n];if(v===undefined||v===null)v='';
var req=f.r?' *':'';
if(f.t==='select'){
h+='<div class="fr"><label>'+f.l+req+'</label><select id="f-'+f.n+'">';
h+='<option value="">Select...</option>';
f.o.forEach(function(opt){h+='<option value="'+opt+'"'+(v===opt?' selected':'')+'>'+opt+'</option>'});
h+='</select></div>';
}else if(f.t==='textarea'){
h+='<div class="fr"><label>'+f.l+req+'</label><textarea id="f-'+f.n+'" rows="3">'+esc(String(v))+'</textarea></div>';
}else if(f.t==='checkbox'){
h+='<div class="fr"><label><input type="checkbox" id="f-'+f.n+'"'+(v?' checked':'')+' style="width:auto;margin-right:.5rem">'+f.l+'</label></div>';
}else{
var inputType='text';
if(f.t==='number'||f.t==='integer')inputType='number';
if(f.t==='email')inputType='email';
if(f.t==='url')inputType='url';
if(f.t==='phone')inputType='tel';
if(f.t==='date')inputType='date';
if(f.t==='datetime')inputType='datetime-local';
var ph=(idx===0&&window._placeholderName&&!v)?' placeholder="'+esc(window._placeholderName)+'"':'';
h+='<div class="fr"><label>'+f.l+req+'</label><input type="'+inputType+'" id="f-'+f.n+'" value="'+esc(String(v))+'"'+ph+'></div>';
}
});
h+='<div class="acts"><button class="btn" onclick="closeModal()">Cancel</button><button class="btn btn-p" onclick="submit()">'+(isEdit?'Save':'Create')+'</button></div>';
return h;
}

document.getElementById('add-btn').onclick=function(){if(window._trialRequired){showTrialNudge();return}editId=null;document.getElementById('mdl').innerHTML=formHTML(activeRes);document.getElementById('mbg').classList.add('open')};
function openEdit(r,id){activeRes=r;var item=null;(data[r]||[]).forEach(function(x){if(x.id===id)item=x});if(!item)return;editId=id;document.getElementById('mdl').innerHTML=formHTML(r,item);document.getElementById('mbg').classList.add('open')}
function closeModal(){document.getElementById('mbg').classList.remove('open');editId=null}

async function submit(){
var r=activeRes,cfg=resCfg[r],body={},extras={};
cfg.f.forEach(function(f){
var el=document.getElementById('f-'+f.n);if(!el)return;
var val;
if(f.t==='checkbox')val=el.checked;
else if(f.t==='number')val=parseFloat(el.value)||0;
else if(f.t==='integer')val=parseInt(el.value)||0;
else val=el.value.trim();
if(f.x)extras[f.n]=val;else body[f.n]=val;
});
for(var j=0;j<cfg.f.length;j++){var f=cfg.f[j];if(f.r&&!f.x&&!body[f.n]){alert(f.l+' is required');return}}
var savedId=editId;
if(editId){await fetch(A+'/'+r+'/'+editId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)})}
else{var resp=await fetch(A+'/'+r,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});if(!resp.ok){var e=await resp.json();alert(e.error||'Error');return}var created=await resp.json();savedId=created.id}
if(Object.keys(extras).length&&savedId){
await fetch(A+'/extras/'+r+'/'+savedId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(extras)});
}
closeModal();loadAll();
}

async function del(r,id){if(!confirm('Delete this record?'))return;await fetch(A+'/'+r+'/'+id,{method:'DELETE'});loadAll()}
function esc(s){if(!s)return'';var d=document.createElement('div');d.textContent=s;return d.innerHTML}
document.addEventListener('keydown',function(e){if(e.key==='Escape')closeModal()});
// Personalization: fetch config.json overrides from /api/config
(function(){
fetch('/api/config').then(function(r){return r.json()}).then(function(cfg){
if(!cfg||!cfg.dashboard_title)return;
var h1=document.querySelector('.hdr h1');
if(h1&&cfg.dashboard_title){h1.innerHTML='<span>&#9670;</span> '+cfg.dashboard_title}
document.title=cfg.dashboard_title||document.title;
if(cfg.custom_fields&&cfg.custom_fields.length){
var firstRes=Object.keys(resCfg)[0];
if(firstRes){
var tbl=document.querySelector('#tab-'+firstRes+' .tbl thead tr');
cfg.custom_fields.forEach(function(f){
resCfg[firstRes].f.push({n:f.name,l:f.label,t:f.type||'text',r:false,o:f.options||[],x:true});
if(tbl){var th=document.createElement('th');th.textContent=f.label;tbl.insertBefore(th,tbl.lastElementChild)}
})}}
if(cfg.empty_state_message){window._emptyMsg=cfg.empty_state_message}
if(cfg.placeholder_name){window._placeholderName=cfg.placeholder_name}
}).catch(function(){}).finally(function(){checkTrialState();loadAll()});
})();

// ─── advanced-tools license gating ───
window._trialRequired=false;

async function checkTrialState(){
try{
var resp=await fetch('/api/tier');
if(!resp.ok)return;
var data=await resp.json();
window._trialRequired=(data.tier==='none'||data.tier==='expired'||data.expired===true);
var bar=document.getElementById('trial-bar');
var msgEl=document.getElementById('trial-bar-text');
if(window._trialRequired){
bar.classList.add('show');
if(msgEl&&data.tier==='expired'){
msgEl.textContent='Your trial or license has expired. Reads work, but writes are locked until you renew.';
}else if(msgEl){
msgEl.textContent='You can view your existing data, but writes are locked until you start a 14-day free trial or activate a license key.';
}
disableWriteControls();
if(typeof renderRes==='function'){renderRes('reservations')}
}else{
bar.classList.remove('show');
}
}catch(e){}
}

function disableWriteControls(){
var btn=document.getElementById('add-btn');
if(btn){
btn.classList.add('btn-disabled-trial');
btn.title='Locked: license required';
}
}

function showTrialNudge(){
var input=document.getElementById('trial-key-input');
if(input){
input.focus();
input.style.borderColor='var(--rust)';
setTimeout(function(){if(input)input.style.borderColor=''},1500);
}
}

async function activateLicense(){
var input=document.getElementById('trial-key-input');
var btn=document.getElementById('trial-activate-btn');
var msg=document.getElementById('trial-msg');
if(!input||!btn||!msg)return;
var key=(input.value||'').trim();
if(!key){
msg.className='trial-msg error';
msg.textContent='Paste your license key first';
input.focus();
return;
}
btn.disabled=true;
msg.className='trial-msg';
msg.textContent='Activating...';
try{
var resp=await fetch('/api/license/activate',{
method:'POST',
headers:{'Content-Type':'application/json'},
body:JSON.stringify({license_key:key})
});
var data=await resp.json();
if(!resp.ok){
msg.className='trial-msg error';
msg.textContent=data.error||'Activation failed';
btn.disabled=false;
return;
}
msg.className='trial-msg success';
msg.textContent='Activated ('+data.tier+'). Reloading...';
setTimeout(function(){location.reload()},800);
}catch(e){
msg.className='trial-msg error';
msg.textContent='Network error: '+e.message;
btn.disabled=false;
}
}

document.addEventListener('DOMContentLoaded',function(){
var input=document.getElementById('trial-key-input');
if(input){
input.addEventListener('keydown',function(e){
if(e.key==='Enter')activateLicense();
});
}
});
</script></body></html>`
