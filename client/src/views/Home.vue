<template>
    <v-container grid-list-md class="home">
    <h1 class="display-3 mb-3">Webhook Scheduler</h1>

    <v-layout row justify-center>
      <v-dialog v-model="dialog" persistent max-width="800px">
        <v-btn slot="activator" color="primary" dark>Schedule New Event</v-btn>
        <v-card>
          <v-card-title justify-center>
            <span class="headline">Schedule a New Event</span>
          </v-card-title>
          <v-card-text>
            <v-container grid-list-md>
              <v-layout wrap>
                <v-flex xs12 md6>
                  <v-date-picker v-model="date" ></v-date-picker>
                </v-flex>
                <v-flex xs12 md6>
                  <v-time-picker v-model="time"></v-time-picker>
                </v-flex>
              </v-layout>
            </v-container>
          </v-card-text>
          <v-card-actions>
            <v-spacer></v-spacer>
            <v-btn color="blue darken-1" flat @click="dialog = false">Cancel</v-btn>
            <v-btn color="blue darken-1" @click="save">Save</v-btn>
          </v-card-actions>
        </v-card>
      </v-dialog>
    </v-layout>

    <v-layout row justify-center>
      <v-flex xs12 sm9 md8>
        <v-container grid-list-md text-xs-center>
          <template v-for="(schedule, index) in schedules">
            <v-layout align-center justify-start :key="schedule.textId">
              <v-flex xs2 :class="{ 'font-weight-bold': schedule.nextEvent }">{{schedule.status}}</v-flex>
              <v-flex xs8 :class="{ 'font-weight-bold': schedule.nextEvent }">{{schedule.timeString}}</v-flex>
              <v-flex xs2>
                <v-btn color="error" flat icon small @click="deleteSchedule(schedule.id)">
                  <v-icon>delete</v-icon>
                </v-btn>
              </v-flex>
            </v-layout>
            <v-divider v-if="index != schedules.length - 1" :key="schedule.divId"></v-divider>
          </template>
        </v-container>
      </v-flex>
    </v-layout>

    <v-layout row class="mt-3">
      <v-flex xs12>
        * All times expressed in local timezone
      </v-flex>
    </v-layout>

    </v-container>
</template>

<script>
// @ is an alias to /src
import moment from 'moment'
export default {
  name: 'home',
  data() {
    return {
      error: "",
      schedules: [],
      dialog: false,
      date: null,
      time: null,
      nextEvent: null,
    }
  },
  methods: {
    async deleteSchedule(id) {
      try {
        const jwt = localStorage.getItem("jwt");
        const res = await fetch(process.env.BASE_URL + "schedules/" + id, {
          method: "DELETE",
          headers: {
              "Authorization": "Bearer " + jwt,
          },
        });
        await this.reloadData();
      } catch(err) {
        alert(err);
      }
    },
    async save() {
      const m = moment(`${this.date}T${this.time}`);
      const payload = `time=${encodeURI(m.format())}`;
      const jwt = localStorage.getItem("jwt");

      try {
        const res = await fetch(process.env.BASE_URL + "schedules", {
          method: "POST",
          headers: {
              "Authorization": "Bearer " + jwt,
              "Content-Type": "application/x-www-form-urlencoded",
          },
          body: payload,
        });
        await this.reloadData();
      } catch(err) {
        alert(err);
      }
      this.dialog = false;
    },

    async reloadData() {
      const jwt = localStorage.getItem("jwt");
      if (jwt == null || jwt == "") {
        this.$router.push("/login");
      }

      try {
        this.nextEvent = null;
        const res = await fetch(process.env.BASE_URL + "schedules", {
            method: "GET",
            headers: {
              "Authorization": "Bearer " + jwt,
            },
        });

        if (res.status == 200) {
          var schedules = await res.json();
          schedules.sort((s1, s2) => -moment(s1.time).diff(moment(s2.time)));
          schedules = schedules.map(s => {
            s.timeString = moment(s.time).format('llll');
            s.divId = 'div-' + s.id
            s.textId = 'text-' + s.id
            if (s.status === "PENDING") {
              this.nextEvent = s
            }
            return s;
          })
          if (this.nextEvent) {
            this.nextEvent.nextEvent = true
          }
          this.schedules = schedules
        } else if (res.status == 401) {
          localStorage.removeItem("jwt");
          this.$router.push("/login");
        } else {
          this.error = "Unexpected error";
          console.log(res);
        }
      } catch(err) {
          console.log("Error!");
          console.log(err);
          this.error = "Unexpected error";
      }
    },
  },
  async mounted() {
    this.date = moment().format("YYYY-MM-DD");
    this.time = moment().format("HH:mm");

    this.reloadData();
  },
}
</script>
