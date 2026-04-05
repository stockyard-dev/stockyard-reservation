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
@media(max-width:600px){.row2{grid-template-columns:1fr}.toolbar{flex-direction:column}.search{min-width:100%}}
</style></head><body>
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
data[k]=r[k]||[];
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
if(!items.length){tbody.innerHTML='<tr><td colspan="'+(cfg.f.length+1)+'" class="empty">No records found.</td></tr>';return}
var h='';items.forEach(function(item){
h+='<tr onclick="openEdit(\''+r+'\',\''+item.id+'\')">';
cfg.f.forEach(function(f){
var v=item[f.n];if(v===undefined||v===null)v='';
if(f.t==='checkbox')v=v?'Yes':'No';
h+='<td>'+esc(String(v))+'</td>';
});
h+='<td><button class="btn btn-d" onclick="event.stopPropagation();del(\''+r+'\',\''+item.id+'\')">&#10005;</button></td>';
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
h+='<div class="fr"><label>'+f.l+req+'</label><input type="'+inputType+'" id="f-'+f.n+'" value="'+esc(String(v))+'"></div>';
}
});
h+='<div class="acts"><button class="btn" onclick="closeModal()">Cancel</button><button class="btn btn-p" onclick="submit()">'+(isEdit?'Save':'Create')+'</button></div>';
return h;
}

document.getElementById('add-btn').onclick=function(){editId=null;document.getElementById('mdl').innerHTML=formHTML(activeRes);document.getElementById('mbg').classList.add('open')};
function openEdit(r,id){activeRes=r;var item=null;(data[r]||[]).forEach(function(x){if(x.id===id)item=x});if(!item)return;editId=id;document.getElementById('mdl').innerHTML=formHTML(r,item);document.getElementById('mbg').classList.add('open')}
function closeModal(){document.getElementById('mbg').classList.remove('open');editId=null}

async function submit(){
var r=activeRes,cfg=resCfg[r],body={};
cfg.f.forEach(function(f){
var el=document.getElementById('f-'+f.n);if(!el)return;
if(f.t==='checkbox')body[f.n]=el.checked;
else if(f.t==='number')body[f.n]=parseFloat(el.value)||0;
else if(f.t==='integer')body[f.n]=parseInt(el.value)||0;
else body[f.n]=el.value.trim();
});
// Check required
for(var j=0;j<cfg.f.length;j++){var f=cfg.f[j];if(f.r&&!body[f.n]){alert(f.l+' is required');return}}
if(editId){await fetch(A+'/'+r+'/'+editId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)})}
else{var resp=await fetch(A+'/'+r,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});if(!resp.ok){var e=await resp.json();alert(e.error||'Error');return}}
closeModal();loadAll();
}

async function del(r,id){if(!confirm('Delete this record?'))return;await fetch(A+'/'+r+'/'+id,{method:'DELETE'});loadAll()}
function esc(s){if(!s)return'';var d=document.createElement('div');d.textContent=s;return d.innerHTML}
document.addEventListener('keydown',function(e){if(e.key==='Escape')closeModal()});
loadAll();
</script></body></html>`
