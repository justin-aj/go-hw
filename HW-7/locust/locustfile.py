import random
from locust import HttpUser, task, constant, tag

ITEMS = [
    {"product_id": "p-001", "quantity": 1, "price": 29.99},
    {"product_id": "p-002", "quantity": 2, "price": 14.50},
    {"product_id": "p-003", "quantity": 1, "price": 99.00},
    {"product_id": "p-004", "quantity": 3, "price": 5.00},
]


def random_order(customer_id):
    return {
        "customer_id": customer_id,
        "items": random.sample(ITEMS, k=random.randint(1, 3)),
    }


class OrderUser(HttpUser):
    wait_time = constant(0)

    def on_start(self):
        self.customer_id = random.randint(1000, 9999)

    @task
    @tag("sync")
    def place_order_sync(self):
        self.client.post(
            "/orders/sync",
            json=random_order(self.customer_id),
            name="/orders/sync",
            timeout=120,
        )

    @task
    @tag("async")
    def place_order_async(self):
        with self.client.post(
            "/orders/async",
            json=random_order(self.customer_id),
            name="/orders/async",
            catch_response=True,
        ) as response:
            if response.status_code == 202:
                response.success()
            else:
                response.failure(
                    "Expected 202, got {}: {}".format(response.status_code, response.text)
                )
