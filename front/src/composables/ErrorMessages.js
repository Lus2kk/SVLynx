

export const ERROR_MESSAGES = {

    'invalid telegram hash signature': 'Ошибка подписи Telegram. Попробуйте снова.',
    'auth data expired, please try again': 'Время авторизации истекло. Начните заново.',
    'the user is not logged in': 'Вы не авторизованы. Войдите в аккаунт.',


    'the session was not found or expired': 'Сессия устарела или не найдена. Обновите страницу.',
    'failed to create a session': 'Не удалось создать сессию. Попробуйте позже.',


    'the email address was entered incorrectly': 'Введите корректный email-адрес.',
    'enter the 6-digit code': 'Введите 6-значный код из письма.',
    'the code has expired, request a new one': 'Код истёк — запросите новый.',
    'invalid code': 'Неверный код. Проверьте письмо и попробуйте снова.',
    'wait 5 seconds before resending the email': 'Подождите 5 секунд перед повторной отправкой.',
    'wait 5 seconds before resending the code': 'Подождите 5 секунд перед повторной попыткой.',
    'exceeded the number of attempts. try again later': 'Слишком много попыток. Повторите позже.',
    'error when sending the code to the mail': 'Не удалось отправить код. Проверьте email или попробуйте позже.',


    'the nickname is already occupied': 'Это имя пользователя уже занято.',
    'the user was not found': 'Пользователь не найден.',
    "couldn't save profile": 'Не удалось сохранить профиль. Попробуйте позже.',


    'invalid request body': 'Некорректный запрос. Проверьте введённые данные.',


    'invalid signature method': 'Ошибка токена. Войдите заново.',
    'invalid token': 'Токен недействителен. Войдите заново.',


    'internal server error': 'Ошибка сервера. Попробуйте позже.',
}


export function translateError(serverMessage) {
    if (!serverMessage) return 'Произошла неизвестная ошибка.'
    const key = serverMessage.trim().toLowerCase()
    return ERROR_MESSAGES[key] ?? serverMessage
}