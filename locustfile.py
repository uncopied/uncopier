from locust import HttpUser, between, task

class WebsiteUser(HttpUser):
    host = "https://uncopied.org"
    wait_time = between(1, 4)

    def on_start(self):
        self.client.post("/api/v1.0/auth/login", "{\"username\":\"tutu\",\"password\":\"tutu\"}")

    @task
    def index(self):
        self.client.get("/")


    @task
    def about(self):
        self.client.get("/static/media/logo.8f8028b5.svg")
