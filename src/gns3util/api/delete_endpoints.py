from . import GNS3APIClient


class GNS3DeleteAPI(GNS3APIClient):

    # user endpoint

    def delete_user(self, user_id):
        return self._api_call(f"access/users/{user_id}", method="DELETE")
