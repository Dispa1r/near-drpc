ELEMENT.locale(ELEMENT.lang.en);
document.title = window.customTitle;

var app = new Vue({
  el:"#app",
  data() {
    return{
      tableData: [],
      selectedData: [],
      total_peers: 0,
      total_rpc_backends: 0,
      total_rpc_nodes: 0,
      timer: null,
      copyText: 'copy',
    }
  },
  computed: {
    endpoint() {
      return window.rpc_url
    }
  },
  mounted() {
    this.getStat()
    this.timer = setInterval(()=>{
      this.getStat()
    }, 10000)
    this.$once('hook:beforeDestroy', () => {
      clearInterval(this.timer);
      this.timer = null;
    });
  },
  methods: {
    getStat() {
      request.get('/api/status').then((res)=> {
        this.tableData = res.data.all_rpc_nodes
        this.selectedData = res.data.backend_rpc_nodes
        this.total_peers = res.data.total_peers
        this.total_rpc_backends = res.data.total_rpc_backends
        this.total_rpc_nodes = res.data.total_rpc_nodes
      })
    },
    formatTime(str) {
      return dayjs(str).format('HH:mm:ss') 
    },
    formatAddr(str) {
      const id = str.split(':')[1]
      if (id.length > 10) {
        return id.substring(0, 5) + '...' + id.substring(id.length - 5);
      }
      return id || '';
    },
    handleCopy() {
      navigator.clipboard.writeText(this.endpoint).then(() => {
        this.copyText = 'copied';
        setTimeout(() => {
          this.copyText = 'copy';
        }, 1000);
      });
    }
  }
})

