<template>
    <v-form>
        <v-container>
            <v-layout row justify-center>
                <v-flex xs12>
                    <h1 align="center">Login</h1>
                </v-flex>
            </v-layout>
            <v-layout row justify-center>
                <v-flex xs12 sm6 md4>
                    <v-text-field
                        label='Email'
                        v-model="email"
                        browser-autocomplete='email'
                        required
                    >
                    </v-text-field>
                </v-flex>
            </v-layout>
            <v-layout row justify-center>
                <v-flex xs12 sm6 md4>
                    <v-text-field
                        label='Password'
                        v-model="password"
                        :append-icon="showPassword ? 'visibility_off' : 'visibility'"
                        :type="showPassword ? 'text' : 'password'"
                        browser-autocomplete='password'
                        required
                        @click:append="showPassword = !showPassword"
                    >
                    </v-text-field>
                </v-flex>
            </v-layout>
            <v-layout red--text v-if="error" row justify-center>
                <v-flex xs12>
                    {{ error }}
                </v-flex>
            </v-layout>
            <v-layout row justify-center>
                <v-flex xs12 sm6 md4>
                    <v-btn @click="submit">submit</v-btn>
                </v-flex>
            </v-layout>
        </v-container>
    </v-form>
</template>

<script>
export default {
    data() {
        return {
            showPassword: false,
            password: "",
            email: "",
            error: "",
        }
    },
    methods: {
        async submit() {
            if (this.email == "" || this.password == "") {
                return
            }

            const body = "email=" + encodeURI(this.email) + "&password=" + encodeURI(this.password)

            try {
                const res = await fetch(process.env.BASE_URL + "login", {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/x-www-form-urlencoded",
                    },
                    body,
                })

                if (res.status == 200) {
                    const jwt = await res.text()
                    localStorage.setItem("jwt", jwt)
                    this.$router.push("/")
                } else if (res.status == 401) {
                    this.error = "Incorrect username or password"
                } else {
                    this.error = "Unexpected error"
                    console.log(res)
                }
            } catch(err) {
                console.log("Error!")
                console.log(err)
                this.error = "Unexpected error"
            }
        }
    }
}
</script>
