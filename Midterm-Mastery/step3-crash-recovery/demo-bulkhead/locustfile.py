from locust import HttpUser, task, between

class DemoUser(HttpUser):
    wait_time = between(0.1, 0.5)

    @task(10)
    def normal_search(self):
        with self.client.get("/products/search?q=beauty", name="[NORMAL] /products/search", catch_response=True) as r:
            if r.status_code == 200:
                r.success()
            else:
                r.failure(f"{r.status_code}")

    @task(3)
    def slow_search(self):
        with self.client.get("/products/search?q=__slow", name="[FAULT] /products/search?q=__slow", catch_response=True) as r:
            if r.status_code in [200, 503, 429]:
                r.success() if r.status_code == 200 else r.failure(f"{r.status_code}")

    @task(2)
    def health_check(self):
        self.client.get("/health")
