<style scoped>
  .action {
    float: right;
  }

  .action .btn {
    border: none;
    margin: 0;
    padding: 0 10px;
    height: 100%;
    opacity: 1;
    background: none;
    color: rgb(0, 96, 242);
    text-transform: uppercase;
    font-size: 16px;
    font-weight: bold;
    text-decoration: none;
  }

  .action .btn:hover {
    color: rgb(126, 211, 33);
  }

  .action .btn.stop:hover {
    color: red;
  }

  .action .btn[disabled],
  .action .btn[disabled]:hover {
    color: rgb(193, 193, 193);
    cursor: not-allowed;
  }

  .row {
    padding-bottom: 10px;
  }

  .fields {
    margin-top: 5px;
  }

  .info-box {
    background-color: rgb(249, 251, 255);
    text-align: center;
    padding: 15px;
  }

  .info-box>span {
    text-transform: uppercase;
    font-weight: bold;
  }
</style>

<template lang="pug">
  div
    h4 current chain
    p.info-box
      span {{this.network}}
    h4 bitmark node
      div.action
        button.btn(
          @click="this.startBitmarkd"
          :disabled="!this.bitmarkd.status || this.bitmarkd.status==='started'") Start
        button.btn.stop(
          disabled, @click="this.stopBitmarkd"
          :disabled="!this.bitmarkd.status || this.bitmarkd.status==='stopped'") Stop
        router-link(tag="button", class="btn",to="/node/config") Config
    p.info-box
      status-grid(
        v-if="this.bitmarkdInfo", title="bitmark info", :style='{backgroundColor: "rgb(249, 251, 255)"}'
        :data="this.bitmarkdInfo", sub-align="horizontal")
      span(v-else) {{ (this.bitmarkd.status === 'started') ? 'Bitmarkd info is not available' : ('Bitmarkd is ' + (this.bitmarkd.status || "loading status")) }}

    h4 prooferd node
      div.action
        button.btn(@click="this.startProoferd"
          :disabled="!this.prooferd.status || this.prooferd.status==='started'") Start
        button.btn.stop(disabled, @click="this.stopProoferd"
          :disabled="!this.prooferd.status || this.prooferd.status==='stopped'") Stop
        router-link(tag="button", class="btn", to="/node/config") Config
    p.info-box
      span Prooferd is {{this.prooferd.status || "loading status"}}

    h4 configuration

    status-grid(title="bitmarkd rpc", :data="{chain: network, announce: bitmarkdConfig.chain}")
    status-grid(title="bitmarkd peer", :data="this.bitmarkdPeerData")
    status-grid(title="prooferd peer", :data="this.prooferdPeerData")
</template>

<script>
  import {
    getCookie
  } from "../utils"
  import axios from "axios"

  import statusGrid from "../components/statusGrid.vue"

  export default {

    computed: {
      bitmarkdPeerData() {
        return (this.bitmarkdConfig.peering) ? {
          publickey: this.bitmarkdConfig.peering.public_key,
          broadcast: this.bitmarkdConfig.peering.announce.broadcast[0],
          listen: this.bitmarkdConfig.peering.announce.listen[0]
        } : {}
      },
      prooferdPeerData() {
        return (this.prooferdConfig.peering) ? {
          connect: this.prooferdConfig.peering.connect
        } : {}
      }
    },
    components: {
      "status-grid": statusGrid
    },
    methods: {
      startBitmarkd() {
        this.bitmarkd.status = ""
        axios.post("/api/" + "bitmarkd", {
          option: "start"
        })
      },

      stopBitmarkd() {
        this.bitmarkd.status = ""
        this.bitmarkdInfo = null;
        axios.post("/api/" + "bitmarkd", {
          option: "stop"
        })
      },

      startProoferd() {
        this.prooferd.status = ""
        axios.post("/api/" + "prooferd", {
          option: "start"
        })
      },

      stopProoferd() {
        this.prooferd.status = ""
        axios.post("/api/" + "prooferd", {
          option: "stop"
        })
      },

      fetchBitmarkInfo() {
        if (this.bitmarkd.status === "started") {
          axios.post("/api/" + "bitmarkd", {
            option: "info"
          }).then((resp) => {
            let data = resp.data
            if (data.ok) {
              this.bitmarkdInfo = data.result
            }
          })
        }
      },

      fetchConfig() {
        axios.get("/api/config")
          .then((resp) => {
            let data = resp.data
            if (data.ok) {
              this.bitmarkdConfig = data.result.bitmarkd.data
              this.prooferdConfig = data.result.prooferd.data
            }
          })
          .catch((e) => {
            this.$emit(e.message)
          })
      },

      fetchStatus(serviceName) {
        let service = this[serviceName]
        if (service.querying) {
          return
        }
        service.querying = true
        axios.post("/api/" + serviceName, {
            option: "status"
          })
          .then((resp) => {
            if (resp.data.ok) {
              service.status = resp.data.result
            } else {
              throw new Error(resp.data.result)
            }
          }).catch((e) => {
            this.$emit("error", e.message)
          })
          .then(() => {
            service.querying = false
          })
      }
    },

    mounted() {
      let network = getCookie("bitmark-webgui-network")
      if (!network) {
        this.$router.push("/chain")
      }
      this.network = network;
      this.fetchConfig()
      this.periodicalTask = setInterval(() => {
        this.fetchStatus('bitmarkd')
        this.fetchStatus('prooferd')
        this.fetchBitmarkInfo()
      }, 2000)
    },

    destroyed() {
      clearInterval(this.periodicalTask)
    },

    data() {
      return {
        network: "",
        periodicalTask: null,
        bitmarkd: {
          querying: false,
          status: ""
        },
        prooferd: {
          querying: false,
          status: ""
        },
        bitmarkdInfo: null,
        bitmarkdConfig: {},
        prooferdConfig: {}
      }
    }
  }
</script>
