from locust import FastHttpUser, task, between
import json
import random


class ProductUser(FastHttpUser):
    """
    A user class that simulates interactions with the Product API.
    Uses FastHttpUser (C-based HTTP client) for better performance.
    """

    # Target host (your Go server)
    host = "http://3.89.25.212:8080"

    # Wait time between tasks (simulates user "think time")
    wait_time = between(0, 0)

    def on_start(self):
        """
        Called when a simulated user starts.
        Seed some products so GET requests can find them.
        """
        self.product_counter = 0
        # Pre-create some products so GETs have data to fetch
        for i in range(1, 11):
            product = {
                "product_id": i,
                "sku": f"SKU-{i:04d}",
                "manufacturer": f"Manufacturer-{i}",
                "category_id": random.randint(1, 50),
                "weight": random.randint(100, 5000),
                "some_other_id": random.randint(1, 1000)
            }
            self.client.post(
                f"/products/{i}/details",
                data=json.dumps(product),
                headers={"Content-Type": "application/json"}
            )

    @task(3)  # Weight of 3 - runs 3x more often than POST
    def get_product(self):
        """
        GET request to retrieve a product by ID.
        This represents read-heavy e-commerce browsing.
        """
        product_id = random.randint(1, 10)

        with self.client.get(
            f"/products/{product_id}",
            name="GET /products/:id",
            catch_response=True
        ) as response:
            if response.status_code == 200:
                try:
                    product = response.json()
                    if product.get("product_id") == product_id:
                        response.success()
                    else:
                        response.failure("Product ID doesn't match")
                except json.JSONDecodeError:
                    response.failure("Invalid JSON response")
            else:
                response.failure(f"Got status code {response.status_code}")

    @task(1)  # Weight of 1 - runs 1x (less frequently)
    def post_product(self):
        """
        POST request to add/update product details.
        This represents a write operation.
        """
        self.product_counter += 1
        product_id = random.randint(11, 10000)
        product = {
            "product_id": product_id,
            "sku": f"SKU-{product_id:04d}-{random.randint(1000, 9999)}",
            "manufacturer": f"Manufacturer-{random.randint(1, 100)}",
            "category_id": random.randint(1, 50),
            "weight": random.randint(100, 5000),
            "some_other_id": random.randint(1, 1000)
        }

        with self.client.post(
            f"/products/{product_id}/details",
            name="POST /products/:id/details",
            data=json.dumps(product),
            headers={"Content-Type": "application/json"},
            catch_response=True
        ) as response:
            if response.status_code == 204:
                response.success()
            else:
                response.failure(f"Got status code {response.status_code}")


class GetOnlyUser(FastHttpUser):
    """
    A user class that only performs GET requests.
    Use this to test read-heavy scenarios.
    """
    host = "http://3.89.25.212:8080"
    wait_time = between(0, 0)

    def on_start(self):
        # Seed products
        for i in range(1, 11):
            product = {
                "product_id": i,
                "sku": f"SKU-{i:04d}",
                "manufacturer": f"Manufacturer-{i}",
                "category_id": random.randint(1, 50),
                "weight": random.randint(100, 5000),
                "some_other_id": random.randint(1, 1000)
            }
            self.client.post(
                f"/products/{i}/details",
                data=json.dumps(product),
                headers={"Content-Type": "application/json"}
            )

    @task
    def get_product(self):
        """Only GET products - read-only test."""
        product_id = random.randint(1, 10)
        self.client.get(f"/products/{product_id}", name="GET /products/:id (read-only)")


class PostOnlyUser(FastHttpUser):
    """
    A user class that only performs POST requests.
    Use this to test write-heavy scenarios.
    """
    host = "http://3.89.25.212:8080"
    wait_time = between(0, 0)

    def on_start(self):
        self.product_counter = 0

    @task
    def post_product(self):
        """Only POST products - write-only test."""
        self.product_counter += 1
        product_id = self.product_counter + 10000
        product = {
            "product_id": product_id,
            "sku": f"SKU-WRITE-{product_id}",
            "manufacturer": "Load Test Corp",
            "category_id": 1,
            "weight": 500,
            "some_other_id": 1
        }
        self.client.post(
            f"/products/{product_id}/details",
            name="POST /products/:id/details (write-only)",
            data=json.dumps(product),
            headers={"Content-Type": "application/json"}
        )